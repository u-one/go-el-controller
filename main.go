package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/u-one/go-el-controller/class"
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
			"ip", "type",
		},
	)
)

func init() {
	prometheus.MustRegister(tempMetrics)

}

var (
	// MulticastIP is Echonet-Lite multicast address
	MulticastIP = "224.0.23.0"
	// Port is Echonet-Lite receive port
	Port = ":3610"

	srv http.Server
)
var (
	// ClassInfoMap is a map with ClassCode as key and PropertyDefs as value
	ClassInfoMap class.ClassInfoMap
)

type ELController struct {
	MulticastReceiver MulticastReceiver
	MulticastSender   MulticastSender
}

func (elc ELController) readMulticast(ctx context.Context, wg sync.WaitGroup) {
	go func() {
		defer wg.Done()

		ch := elc.MulticastReceiver.Start(ctx)

		handler := func(results <-chan ReceiveResult) {
			for result := range results {
				if result.Err != nil {
					log.Println(result.Err)
					continue
				}
				frame, err := ParseFrame(result.Data)
				if err != nil {
					log.Println(err)
					continue
				}
				log.Printf("[%v] %v\n", result.Address, frame)
				err = frameReceived(result.Address, frame)
				if err != nil {
					log.Println(err)
				}
			}
		}

		handler(ch)
	}()
}

func (elc ELController) readUnicast(ctx context.Context, wg sync.WaitGroup) {
}

func frameReceived(addr string, f Frame) error {
	//sObjInfo := ClassInfoMap.Get(f.SrcClassCode)
	classCode := f.SrcClassCode()
	log.Println("frameReceived:", classCode)
	f.Print()

	switch ClassGroupCode(classCode.ClassGroupCode) {
	case AirConditioner:
		switch ClassCode(classCode.ClassCode) {
		case HomeAirConditioner:
			log.Println("エアコン")
			for _, p := range f.Properties {
				log.Println(p.Code)
				switch PropertyCode(p.Code) {
				case MeasuredRoomTemperature:
					if p.Len != 1 {
						return fmt.Errorf("invalid length: %d", p.Len)
					}
					temp := int(p.Data[0])
					log.Printf("室温:%d℃\n", temp)
					tempMetrics.With(prometheus.Labels{"ip": addr, "type": "room"}).Set(float64(temp))
					break
				case MeasuredOutdoorTemperature:
					if p.Len != 1 {
						return fmt.Errorf("invalid length: %d", p.Len)
					}
					temp := int(p.Data[0])
					log.Printf("外気温:%d℃\n", temp)
					tempMetrics.With(prometheus.Labels{"ip": addr, "type": "outside"}).Set(float64(temp))
					break
				}
			}
			break
		}
		break
	}
	return nil
}

func start() {
	ctx := context.Background()
	//ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup

	ms, err := NewUDPMulticastSender(MulticastIP, Port)
	if err != nil {
		log.Println(err)
		return
	}
	defer ms.Close()

	elc := ELController{
		MulticastReceiver: &UDPMulticastReceiver{
			IP:   MulticastIP,
			Port: Port,
		},
		MulticastSender: ms,
	}

	wg.Add(1)
	elc.readMulticast(ctx, wg)

	//wg.Add(1)
	//elc.readUnicast(ctx, wg)

	wg.Add(1)
	elc.sendLoop(ctx, wg)
	//f = createAirconGetFrame()
	//sendFrame(conn, f)

	wg.Add(1)
	startExporter(ctx, wg)

	log.Println("wait for read done")
	wg.Wait()
	log.Println("finish ")
}

func (elc ELController) sendLoop(ctx context.Context, wg sync.WaitGroup) {
	go func() {
		defer wg.Done()

		sendFrame := func(f *Frame) {
			log.Println("sendFrame")
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
		t.Stop()
	}()
}

var exporterAddr = flag.String("listen-address", ":8083", "The address to listen on for HTTP requests.")

func startExporter(ctx context.Context, wg sync.WaitGroup) {

	ch := make(chan error)
	go func() {
		defer close(ch)
		srv := http.NewServeMux()
		srv.Handle("/metrics", promhttp.Handler())
		defer wg.Done()
		log.Println("startExporter: ", *exporterAddr)
		http.Handle("/metrics", promhttp.Handler())
		select {
		case ch <- http.ListenAndServe(*exporterAddr, srv):
		case <-ctx.Done():
		}
		log.Println("exporter finished")
	}()
}

func main() {
	flag.Parse()

	ClassInfoMap = class.Load()

	start()

}
