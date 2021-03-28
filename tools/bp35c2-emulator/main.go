package main

import (
	"flag"
	"log"

	"github.com/u-one/go-el-controller/wisun"
)

var portAddr = flag.String("port", "COM4", "Serial port address (COM4)")

func main() {
	log.Println("Started port:", *portAddr)
	e := wisun.NewBP35C2Emulator(*portAddr)
	defer e.Close()
	e.Start()
}
