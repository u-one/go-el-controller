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
)

var version string

var exporterAddr = flag.String("listen-address", ":8083", "The address to listen on for HTTP requests.")

var (
	verCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "home",
			Subsystem: "elexporter",
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

	fmt.Printf("version: %s\n", version)
	verCounter.WithLabelValues(version).Inc()

	err := echonetlite.PrepareClassDictionary()
	if err != nil {
		log.Println(err)
	}

	ctx := context.Background()
	//ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch := make(chan error)
	go func() {
		defer close(ch)
		server := http.NewServeMux()
		server.Handle("/metrics", promhttp.Handler())
		log.Println("startExporter: ", *exporterAddr)
		http.Handle("/metrics", promhttp.Handler())
		select {
		case ch <- http.ListenAndServe(*exporterAddr, server):
		case <-ctx.Done():
		}
		log.Println("exporter finished")
	}()

	elc, err := echonetlite.NewControllerNode()
	if err != nil {
		log.Println(err)
		return
	}
	elc.Start(ctx)
	defer elc.Close()

	log.Println("start sendLoop")

	func() {
		t := time.NewTicker(30 * time.Second)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				elc.RequestAirConState()
			case <-ctx.Done():
				return
			}
		}
	}()

	log.Println("finished")
}
