package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/u-one/go-el-controller/echonetlite"
	"github.com/u-one/go-el-controller/transport"
)

var clogger *log.Logger

func init() {
	clogger = log.New(os.Stdout, "[Controller]", log.LstdFlags)
}

var (
	tempMetrics = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "home",
			Subsystem: "aircon",
			Name:      "temperature",
			Help:      "aircon temp",
		},
		[]string{
			"ip", "type", "location",
		},
	)
)

func init() {
	prometheus.MustRegister(tempMetrics)

}

const (
	// MulticastIP is Echonet-Lite multicast address
	MulticastIP = "224.0.23.0"
	// Port is Echonet-Lite receive port
	Port = ":3610"
)

// ELController is ECHONETLite controller
type ELController struct {
	MulticastReceiver transport.MulticastReceiver
	MulticastSender   transport.MulticastSender
	ExporterAddr      string
	Server            *http.ServeMux
}

// Start starts controller
func (elc ELController) Start(ctx context.Context) {
	elc.readMulticast(ctx)

	//wg.Add(1)
	//elc.readUnicast(ctx, wg)

	elc.sendLoop(ctx)
	//f = createAirconGetFrame()
	//sendFrame(conn, f)

	elc.startExporter(ctx)

	//clogger.Println("wait for read done")
	//wg.Wait()
	//clogger.Println("finish ")

}

func (elc ELController) readMulticast(ctx context.Context) {
	go func() {

		ch := elc.MulticastReceiver.Start(ctx)

		handler := func(results <-chan transport.ReceiveResult) {
			for {
				select {
				case <-ctx.Done():
					clogger.Println("readMulticast handler ctx.Done")
					return
				case result := <-results:
					if result.Err != nil {
						clogger.Printf("[Error] failed to receive [%s]", result.Err)
						break
					}
					clogger.Println("<<<<<<<< received")
					frame, err := echonetlite.ParseFrame(result.Data)
					if err != nil {
						clogger.Printf("[Error] parse failed [%s]", err)
						break
					}
					clogger.Printf("[%v] %v\n", result.Address, frame)

					switch obj := frame.Object.(type) {
					case echonetlite.AirconObject:
						lc := obj.InstallLocation.Code
						ln := obj.InstallLocation.Number
						loc := lc.String()
						if ln != 0 {
							loc = fmt.Sprintf("%s%d", lc, ln)
						}
						tempMetrics.With(prometheus.Labels{"ip": result.Address, "location": loc, "type": "room"}).Set(obj.InternalTemp)
						tempMetrics.With(prometheus.Labels{"ip": result.Address, "location": loc, "type": "outside"}).Set(obj.OuterTemp)
					default:
					}
				}
			}
		}
		handler(ch)
	}()
}

func (elc ELController) readUnicast(ctx context.Context) {
}

func (elc ELController) sendLoop(ctx context.Context) {
	go func() {

		sendFrame := func(f *echonetlite.Frame) {
			clogger.Println(">>>>>>>> sendFrame")
			f.Print()
			elc.MulticastSender.Send([]byte(f.Data))
		}

		f := echonetlite.CreateInfFrame()
		sendFrame(f)

		// ver.1.0
		f = echonetlite.CreateInfReqFrame()
		sendFrame(f)

		// ver.1.1
		f = echonetlite.CreateGetFrame()
		sendFrame(f)

		time.Sleep(time.Second * 3)

		t := time.NewTicker(30 * time.Second)
		defer t.Stop()
		f = echonetlite.CreateAirconGetFrame()
		sendFrame(f)

		clogger.Println("start sendLoop")

		for {
			select {
			case <-t.C:
				f := echonetlite.CreateAirconGetFrame()
				sendFrame(f)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (elc ELController) startExporter(ctx context.Context) {

	ch := make(chan error)
	go func() {
		defer close(ch)
		elc.Server = http.NewServeMux()
		elc.Server.Handle("/metrics", promhttp.Handler())
		clogger.Println("startExporter: ", elc.ExporterAddr)
		http.Handle("/metrics", promhttp.Handler())
		select {
		case ch <- http.ListenAndServe(elc.ExporterAddr, elc.Server):
		case <-ctx.Done():
		}
		clogger.Println("exporter finished")
	}()
}
