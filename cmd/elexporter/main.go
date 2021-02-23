package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/u-one/go-el-controller/echonetlite"
)

var version string

var exporterAddr = flag.String("listen-address", ":8083", "The address to listen on for HTTP requests.")

func main() {
	flag.Parse()

	fmt.Printf("version: %s\n", version)

	var err error
	// TODO: refactor
	echonetlite.ClassInfoDB, err = echonetlite.Load()
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
		clogger.Println("startExporter: ", *exporterAddr)
		http.Handle("/metrics", promhttp.Handler())
		select {
		case ch <- http.ListenAndServe(*exporterAddr, server):
		case <-ctx.Done():
		}
		clogger.Println("exporter finished")
	}()

	elc, err := NewELController()
	if err != nil {
		log.Println(err)
		return
	}
	elc.Start(ctx)
	defer elc.Close()

	clogger.Println("start sendLoop")

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
