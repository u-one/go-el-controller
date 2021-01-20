package main

import (
	"flag"

	"github.com/u-one/go-el-controller/hems"
	"github.com/u-one/go-el-controller/wisun"
)

var bRouteID = flag.String("brouteid", "", "B-route ID")
var bRoutePW = flag.String("broutepw", "", "B-route password")

func main() {
	flag.Parse()

	wisunClient := wisun.NewBP35C2Client()
	emc := hems.NewElectricityMeterClient(wisunClient)
	defer emc.Close()

	emc.Start(*bRouteID, *bRoutePW)
}
