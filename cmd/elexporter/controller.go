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
	UnicastReceiver   transport.UnicastReceiver
	MulticastSender   transport.MulticastSender
	ExporterAddr      string
	Server            *http.ServeMux
	tid               uint16
	nodeList          NodeList
}

// NewELController returns ELController
func NewELController(exporterAddr string) (*ELController, error) {
	ms, err := transport.NewUDPMulticastSender(MulticastIP, Port)
	if err != nil {
		log.Println(err)
		return &ELController{}, err
	}
	return &ELController{
		MulticastReceiver: &transport.UDPMulticastReceiver{},
		MulticastSender:   ms,
		UnicastReceiver:   &transport.UDPUnicastReceiver{},
		ExporterAddr:      exporterAddr,
	}, nil
}

// Close closes all resources open
func (elc ELController) Close() {
	elc.MulticastSender.Close()
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

	sch := elc.UnicastReceiver.Start(ctx, Port)
	go elc.handleResult(ctx, sch)

	mch := elc.MulticastReceiver.Start(ctx, MulticastIP, Port)
	go elc.handleMulticastResult(ctx, mch)

	elc.startSequence(ctx)
	elc.sendLoop(ctx)

	elc.startExporter(ctx)

	//clogger.Println("wait for read done")
	//wg.Wait()
	//clogger.Println("finish ")

}

func (elc ELController) handleMulticastResult(ctx context.Context, results <-chan transport.ReceiveResult) {
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

func (elc ELController) handleResult(ctx context.Context, results <-chan transport.ReceiveResult) {
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

func (elc *ELController) sendFrame(f *echonetlite.Frame) {
	clogger.Printf(">>>>>>>> SEND : %s\n", f)
	elc.MulticastSender.Send([]byte(f.Serialize()))
	elc.tid++
}

func (elc *ELController) startSequence(ctx context.Context) {
	f := echonetlite.CreateInfFrame(elc.tid)
	elc.sendFrame(f)

	// ver.1.0
	f = echonetlite.CreateInfReqFrame(elc.tid)
	elc.sendFrame(f)

	// ver.1.1
	f = echonetlite.CreateGetFrame(elc.tid)
	elc.sendFrame(f)

	time.Sleep(time.Second * 3)
}

func (elc ELController) sendLoop(ctx context.Context) {
	clogger.Println("start sendLoop")

	t := time.NewTicker(30 * time.Second)
	defer t.Stop()
	f := echonetlite.CreateAirconGetFrame(elc.tid)
	elc.sendFrame(f)

	for {
		select {
		case <-t.C:
			elc.sendFrame(f)
		case <-ctx.Done():
			return
		}
	}
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
