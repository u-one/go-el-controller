package main

import (
	"log"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func init() {
	log.SetOutput(os.Stdout)
}

//[192.168.1.10:40929] 108100390ef0010ef0016201d600
//[192.168.1.10:49322] 108100410ef0010ef0017301d5040105ff01
//[192.168.1.17:4527] 108100050130010ef0017301800130
//[192.168.1.17:4530] 108100060130010ef0017301800131

func Test_NewFrame(t *testing.T) {
	data := [][]byte{
		{0x10, 0x81, 0x00, 0x39, 0x0e, 0xf0, 0x01, 0x0e, 0xf0, 0x01, 0x62, 0x01, 0xd6, 0x00},
	}

	f, err := NewFrame(data[0])
	if err != nil {
		t.Error(err)
	}
	log.Printf("[%v] %v\n", "", f)

	e := Frame{}

	if diff := cmp.Diff(f, e); diff != "" {
		t.Errorf("Hogefunc differs: (-got +want)\n%s", diff)
	}
}
