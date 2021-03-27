package main

import (
	"flag"

	"github.com/u-one/go-el-controller/wisun"
)

var portAddr = flag.String("port", "COM5", "Serial port address (COM4)")

func main() {
	e := wisun.NewBP35C2Emulator(*portAddr)
	defer e.Close()
	e.Start()
}
