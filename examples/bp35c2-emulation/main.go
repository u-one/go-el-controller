package main

import (
	"github.com/u-one/go-el-controller/wisun"
)

func main() {
	e := wisun.NewBP35C2Emulator("COM3")
	defer e.Close()
	e.Start()
}
