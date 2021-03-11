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
	wisunClient := wisun.NewBP35C2Client()
	defer wisunClient.Close()

	err := wisunClient.Connect(*bRouteID, *bRoutePW)
	if err != nil {
		return fmt.Errorf("Connect failed: %v", err)
	}

	f := echonetlite.CreateCurrentPowerConsumptionFrame(1)

	eldata, err := wisunClient.Send(f.Serialize())
	if err != nil {
		return err
	}
	elFrame, err := echonetlite.ParseFrame(eldata)
	if err != nil {
		return fmt.Errorf("invalid frame: %w", err)
	}
	elFrame.Print()

	return nil
}
