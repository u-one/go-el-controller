package main

import (
	"context"
	"flag"
	"log"

	"github.com/u-one/go-el-controller/class"
)

var (
	// MulticastIP is Echonet-Lite multicast address
	MulticastIP = "224.0.23.0"
	// Port is Echonet-Lite receive port
	Port = ":3610"
)
var (
	// ClassInfoMap is a map with ClassCode as key and PropertyDefs as value
	ClassInfoMap class.ClassInfoMap
)

func start() {
	ctx := context.Background()
	//ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

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
		ExporterAddr:    *exporterAddr,
	}

	elc.Start(ctx)

}

var exporterAddr = flag.String("listen-address", ":8083", "The address to listen on for HTTP requests.")

func main() {
	flag.Parse()

	var err error
	ClassInfoMap, err = class.Load()
	if err != nil {
		log.Println(err)
	}

	start()

}
