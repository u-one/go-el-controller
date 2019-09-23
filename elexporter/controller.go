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
	tid               uint16
	nodeList          NodeList
}

// NodeList is list of node profile objects
type NodeList map[string]Node

// Add adds Node
func (nlist *NodeList) Add(addr string, obj echonetlite.Object) {
	if _, ok := (*nlist)[addr]; !ok {
		(*nlist)[addr] = Node{}
	}
}

// Node represents a node profile object
type Node struct {
	Devices []echonetlite.Object
}

// Start starts controller
func (elc ELController) Start(ctx context.Context) {
	elc.tid = 0
	elc.nodeList = make(NodeList)

	elc.readMulticast(ctx)

	//wg.Add(1)
	//elc.readUnicast(ctx)

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

		ch := elc.MulticastReceiver.Start(ctx, MulticastIP, Port)

		handler := func(results <-chan transport.ReceiveResult) {
			for {
				select {
				case <-ctx.Done():
					clogger.Println("readMulticast handler ctx.Done")
					return
				case result := <-results:
					if result.Err != nil {
						clogger.Printf("[Error] failed to receive [%s]\n", result.Err)
						break
					}
					clogger.Printf("<<<<<<<< [%v] MULTI CAST RECEIVE: ", result.Address)
					frame, err := echonetlite.ParseFrame(result.Data)
					if err != nil {
						clogger.Printf("[Error] parse failed [%s]\n", err)
						break
					}
					clogger.Printf("[%v] %s\n", result.Address, frame)

					switch frame.ESV {
					case echonetlite.Inf:
						elc.nodeList.Add(result.Address, frame.SEOJ)
						//[Controller]2019/09/27 01:52:59 [192.168.1.15] 108100010ef00105ff017301d50401013001 EHD[1081] TID[0001] SEOJ[0ef001](ノードプロファイル) DEOJ[05ff01](コントローラ) ESV[INF] OPC[01] EPC0[d5](インスタンスリスト通知) PDC0[4] EDT0[01013001]
						//[Controller]2019/09/27 01:52:59 [192.168.1.10] 108100010ef00105ff017301d50401013001 EHD[1081] TID[0001] SEOJ[0ef001](ノードプロファイル) DEOJ[05ff01](コントローラ) ESV[INF] OPC[01] EPC0[d5](インスタンスリスト通知) PDC0[4] EDT0[01013001]

					default:
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
		}
		handler(ch)
	}()
}

func (elc ELController) readUnicast(ctx context.Context) {
	/*
		go func() {
			ch := elc.UnicastReceiver.Start(ctx, "localhost", ":3611")

			handler := func(results <-chan transport.ReceiveResult) {
				for {
					select {
					case <-ctx.Done():
						clogger.Println("readUnicast handler ctx.Done")
						return
					case result := <-results:
						if result.Err != nil {
							clogger.Printf("[Error] failed to receive [%s]\n", result.Err)
							break
						}
						clogger.Printf("<<<<<<<< [%v] UNI CAST RECEIVE: ", result.Address)
						frame, err := echonetlite.ParseFrame(result.Data)
						if err != nil {
							clogger.Printf("[Error] parse failed [%s]\n", err)
							break
						}
						clogger.Printf("[%v] %v\n", result.Address, frame)
					}
				}
			}
			handler(ch)
		}()
	*/

}

func (elc ELController) sendLoop(ctx context.Context) {
	go func() {

		sendFrame := func(f *echonetlite.Frame) {
			clogger.Printf(">>>>>>>> SEND : %s\n", f)
			elc.MulticastSender.Send([]byte(f.Data))
			elc.tid++
		}

		f := echonetlite.CreateInfFrame(elc.tid)
		sendFrame(f)

		// ver.1.0
		f = echonetlite.CreateInfReqFrame(elc.tid)
		sendFrame(f)

		// ver.1.1
		//f = echonetlite.CreateGetFrame()
		//sendFrame(f)

		time.Sleep(time.Second * 3)

		t := time.NewTicker(30 * time.Second)
		defer t.Stop()
		f = echonetlite.CreateAirconGetFrame(elc.tid)
		sendFrame(f)

		clogger.Println("start sendLoop")

		for {
			select {
			case <-t.C:
				f := echonetlite.CreateAirconGetFrame(elc.tid)
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
