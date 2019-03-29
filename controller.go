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
			"ip", "type",
		},
	)
)

func init() {
	prometheus.MustRegister(tempMetrics)

}

type ELController struct {
	MulticastReceiver MulticastReceiver
	MulticastSender   MulticastSender
	ExporterAddr      string
	Server            *http.ServeMux
}

func (elc ELController) Start(ctx context.Context) {
	var wg sync.WaitGroup

	wg.Add(1)
	elc.readMulticast(ctx, wg)

	//wg.Add(1)
	//elc.readUnicast(ctx, wg)

	wg.Add(1)
	elc.sendLoop(ctx, wg)
	//f = createAirconGetFrame()
	//sendFrame(conn, f)

	wg.Add(1)
	elc.startExporter(ctx, wg)

	log.Println("wait for read done")
	wg.Wait()
	log.Println("finish ")

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
				err = elc.frameReceived(result.Address, frame)
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

func (elc ELController) frameReceived(addr string, f Frame) error {
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

func (elc ELController) startExporter(ctx context.Context, wg sync.WaitGroup) {

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
