package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/u-one/go-el-controller/echonetlite"
	"github.com/u-one/go-el-controller/wisun"
)

var version string
var bRouteID = flag.String("brouteid", "", "B-route ID")
var bRoutePW = flag.String("broutepw", "", "B-route password")

var serialPort = flag.String("serial-port", "/dev/ttyUSB0", "serial port for BP35C2")
var exporterPort = flag.String("exporter-port", "8080", "address for prometheus")

var (
	verCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "home",
			Subsystem: "smartmeter_exporter",
			Name:      "version",
			Help:      "app version",
		},
		[]string{
			"version",
		},
	)
)

func init() {
	prometheus.MustRegister(verCounter)
}

func main() {
	flag.Parse()
	err := run()
	if err != nil {
		log.Println(err)
	}
}

func run() error {

	fmt.Printf("version: %s serial-port:%s exporter-port:%s\n", version, *serialPort, *exporterPort)
	verCounter.WithLabelValues(version).Inc()

	wisunClient := wisun.NewBP35C2Client(*serialPort)
	defer wisunClient.Close()
	node := echonetlite.NewElectricityControllerNode(wisunClient)
	defer node.Close()

	err := node.Start(*bRouteID, *bRoutePW)
	if err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start prometheus exporter
	ch := make(chan error)
	go func() {
		defer close(ch)
		server := http.NewServeMux()
		server.Handle("/metrics", promhttp.Handler())
		log.Println("start exporter: ", *exporterPort)
		select {
		case ch <- http.ListenAndServe(":"+*exporterPort, server):
		case <-ctx.Done():
		}
		log.Println("exporter finished")
	}()

	func() {
		t := time.NewTicker(5 * time.Second)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				_, err := node.GetPowerConsumption()
				if err != nil {
					log.Println(err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}
