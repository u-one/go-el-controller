package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/u-one/go-el-controller/hems"
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
	wisunClient := wisun.NewBP35C2Client()
	emc := hems.NewElectricityMeterClient(wisunClient)
	defer emc.Close()

	err := emc.Start(*bRouteID, *bRoutePW)
	if err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	for {
		_, err := emc.GetPowerConsumption()
		if err != nil {
			log.Println(err)
		}
	}
}
