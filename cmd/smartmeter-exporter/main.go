package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/u-one/go-el-controller/echonetlite"
	"github.com/u-one/go-el-controller/wisun"
)

var bRouteID = flag.String("brouteid", "", "B-route ID")
var bRoutePW = flag.String("broutepw", "", "B-route password")

func main() {
	flag.Parse()
	err := run()
	if err != nil {
		log.Println(err)
	}
}

func run() error {
	serialport := "/dev/ttyUSB0"
	wisunClient := wisun.NewBP35C2Client(serialport)
	node := echonetlite.NewElectricityControllerNode(wisunClient)
	defer node.Close()

	err := node.Start(*bRouteID, *bRoutePW)
	if err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	for {
		_, err := node.GetPowerConsumption()
		if err != nil {
			log.Println(err)
		}
	}
}
