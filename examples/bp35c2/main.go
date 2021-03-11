package main

import (
	"flag"
	"fmt"
	"log"

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

	p, err := wisunClient.Get()
	if err != nil {
		return fmt.Errorf("Get failed: %v", err)
	}
	log.Println(p)
	return nil
}
