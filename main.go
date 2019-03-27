package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
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

func readMulticast(ctx context.Context, wg sync.WaitGroup) {
	go func() {
		defer wg.Done()
		log.Println("Start to listen multicast udp ", MulticastIP, Port)
		address, err := net.ResolveUDPAddr("udp", MulticastIP+Port)
		log.Println("resolved:", address)
		if err != nil {
			log.Println("Error: ", err)
			return
		}
		conn, err := net.ListenMulticastUDP("udp", nil, address)
		if err != nil {
			log.Println("Error:", err)
			return
		}
		defer conn.Close()
		buffer := make([]byte, 1500)

		for {
			fmt.Printf(".")
			conn.SetDeadline(time.Now().Add(1 * time.Second))
			length, remoteAddress, err := conn.ReadFromUDP(buffer)
			if err != nil {
				err, ok := err.(net.Error)
				if !ok || !err.Timeout() {
					log.Println("Error: ", err)
				}
			}
			if length > 0 {
				fmt.Println()
				frame, err := NewFrame(buffer[:length])
				if err != nil {
					log.Println(err)
					continue
				}
				log.Printf("[%v] %v\n", remoteAddress.IP.String(), frame)
				err = frameReceived(remoteAddress.IP.String(), frame)
				if err != nil {
					log.Println(err)
				}
			}
			select {
			case <-ctx.Done():
				log.Println("ctx.Done")
				return
			default:
				//log.Println("recv: ", length)
			}
		}
	}()
}

func readUnicast(ctx context.Context, wg sync.WaitGroup) {
	go func() {
		defer wg.Done()

		udpAddr, err := net.ResolveUDPAddr("udp", MulticastIP+Port)
		if err != nil {
			log.Println("Unicast Error: ", err)
			return
		}
		conn, err := net.ListenUDP("udp", udpAddr)
		if err != nil {
			log.Println("Unicast Error: ", err)
			return
		}
		defer conn.Close()

		buffer := make([]byte, 1500)
		for {
			conn.SetDeadline(time.Now().Add(1 * time.Second))
			length, remoteAddress, err := conn.ReadFrom(buffer)
			if err != nil {
				log.Println("Unicast Error:", err)
			}
			if length > 0 {
				fmt.Println()
				frame, err := NewFrame(buffer[:length])
				if err != nil {
					log.Println(err)
					continue
				}
				log.Printf("Unicast [%v] %v\n", remoteAddress, frame)
				err = frameReceived(remoteAddress.String(), frame)
				if err != nil {
					log.Println(err)
				}
			}
			select {
			case <-ctx.Done():
				log.Println("ctx.Done")
				return
			default:
			}
		}
	}()
}

func frameReceived(addr string, f Frame) error {
	//sObjInfo := ClassInfoMap.Get(f.ClassCode)
	classCode := f.ClassCode
	log.Println("frameReceived:", classCode)

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

func sendFrame(conn net.Conn, frame *Frame) {
	log.Println("sendFrame")
	write(conn, []byte(frame.Data))
}

func write(conn net.Conn, data []byte) {

	length, err := conn.Write(data)
	if err != nil {
		log.Println("Write error: ", err)
	}
	log.Println("written:", length)
}

func start() {
	ctx := context.Background()
	//ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(1)
	readMulticast(ctx, wg)

	//wg.Add(1)
	//readUnicast(ctx, wg)

	wg.Add(1)
	sendLoop(ctx, wg)
	//f = createAirconGetFrame()
	//sendFrame(conn, f)

	wg.Add(1)
	startExporter(ctx, wg)

	log.Println("wait for read done")
	wg.Wait()
	log.Println("finish ")
}

func sendLoop(ctx context.Context, wg sync.WaitGroup) {
	go func() {
		defer wg.Done()
		conn, err := net.Dial("udp", MulticastIP+Port)
		if err != nil {
			log.Println("Write conn error:", err)
			return
		}
		defer conn.Close()

		f := createInfFrame()
		sendFrame(conn, f)

		// ver.1.0
		f = createInfReqFrame()
		sendFrame(conn, f)

		// ver.1.1
		f = createGetFrame()
		sendFrame(conn, f)

		time.Sleep(time.Second * 3)

		t := time.NewTicker(30 * time.Second)
		f = createAirconGetFrame()
		sendFrame(conn, f)

		log.Println("start sendLoop")

		for {
			select {
			case <-t.C:
				f := createAirconGetFrame()
				sendFrame(conn, f)
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
