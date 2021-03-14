package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/u-one/go-el-controller/echonetlite"
	"github.com/u-one/go-el-controller/wisun"
)

var bRouteID = flag.String("brouteid", "0123456789AB", "B-route ID")
var bRoutePW = flag.String("broutepw", "00112233445566778899AABBCCDDEEFF", "B-route password")

func main() {
	flag.Parse()
	err := run()
	if err != nil {
		log.Println(err)
	}
}

func run() error {
	serialport := "COM2"
	wisunClient := wisun.NewBP35C2Client(serialport)
	defer wisunClient.Close()

	err := wisunClient.Version()
	if err != nil {
		return fmt.Errorf("failed to exec Version: %v", err)
	}

	err = wisunClient.Connect(*bRouteID, *bRoutePW)
	if err != nil {
		return fmt.Errorf("failed to exec Connect: %v", err)
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
