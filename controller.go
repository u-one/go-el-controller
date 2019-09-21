package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

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

// ELController is ECHONETLite controller
type ELController struct {
	MulticastReceiver MulticastReceiver
	MulticastSender   MulticastSender
	ExporterAddr      string
	Server            *http.ServeMux
}

// Start starts controller
func (elc ELController) Start(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(1)
	elc.readMulticast(ctx, &wg)

	//wg.Add(1)
	//elc.readUnicast(ctx, wg)

	wg.Add(1)
	elc.sendLoop(ctx, &wg)
	//f = createAirconGetFrame()
	//sendFrame(conn, f)

	wg.Add(1)
	elc.startExporter(ctx, &wg)

	log.Println("wait for read done")
	wg.Wait()
	log.Println("finish ")

}

func (elc ELController) readMulticast(ctx context.Context, wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()

		ch := elc.MulticastReceiver.Start(ctx)

		handler := func(results <-chan ReceiveResult) {
			for result := range results {
				if result.Err != nil {
					log.Printf("[Error] failed to receive [%s]", result.Err)
					continue
				}
				log.Println("<<<<<<<< received")
				frame, err := ParseFrame(result.Data)
				if err != nil {
					log.Printf("[Error] parse failed [%s]", err)
					continue
				}
				log.Printf("[%v] %v\n", result.Address, frame)

				switch obj := frame.Object.(type) {
				case AirconObject:
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
		handler(ch)
	}()
}

func (elc ELController) readUnicast(ctx context.Context, wg *sync.WaitGroup) {
}

func (elc ELController) sendLoop(ctx context.Context, wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()

		sendFrame := func(f *Frame) {
			log.Println(">>>>>>>> sendFrame")
			f.Print()
			elc.MulticastSender.Send([]byte(f.Data))
		}

		f := createInfFrame()
		sendFrame(f)

		// ver.1.0
		f = createInfReqFrame()
		sendFrame(f)

		// ver.1.1
		f = createGetFrame()
		sendFrame(f)

		time.Sleep(time.Second * 3)

		t := time.NewTicker(30 * time.Second)
		defer t.Stop()
		f = createAirconGetFrame()
		sendFrame(f)

		log.Println("start sendLoop")

		for {
			select {
			case <-t.C:
				f := createAirconGetFrame()
				sendFrame(f)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (elc ELController) startExporter(ctx context.Context, wg *sync.WaitGroup) {

	ch := make(chan error)
	go func() {
		defer close(ch)
		elc.Server = http.NewServeMux()
		elc.Server.Handle("/metrics", promhttp.Handler())
		defer wg.Done()
		log.Println("startExporter: ", elc.ExporterAddr)
		http.Handle("/metrics", promhttp.Handler())
		select {
		case ch <- http.ListenAndServe(elc.ExporterAddr, elc.Server):
		case <-ctx.Done():
		}
		log.Println("exporter finished")
	}()
}
