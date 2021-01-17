package main

import (
	"flag"

	"github.com/u-one/go-el-controller/wisun"
)

var bRouteID = flag.String("brouteid", "", "B-route ID")
var bRoutePW = flag.String("broutepw", "", "B-route password")

func main() {
	flag.Parse()

	wisunClient := wisun.NewBP35C2Client()
	c := wisun.NewElectricityMeterClient(wisunClient)
	defer c.Close()

	c.StartSequence(*bRouteID, *bRoutePW)
}
