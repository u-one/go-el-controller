package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/u-one/go-el-controller/echonetlite"
	"github.com/u-one/go-el-controller/transport"
)

var version string

func start() {
	ctx := context.Background()
	//ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ms, err := transport.NewUDPMulticastSender(MulticastIP, Port)
	if err != nil {
		log.Println(err)
		return
	}
	defer ms.Close()

	elc := ELController{
		MulticastReceiver: &transport.UDPMulticastReceiver{},
		MulticastSender:   ms,
		ExporterAddr:      *exporterAddr,
	}

	elc.Start(ctx)

	select {
	case <-ctx.Done():
		log.Println("finished")
	}
}

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

	start()

}
