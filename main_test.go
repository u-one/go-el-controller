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

func Test_ParseFrame(t *testing.T) {
	data := [][]byte{
		{0x10, 0x81, 0x0, 0x0, 0x05, 0xff, 0x01, 0x01, 0x30, 0x01, 0x62, 0x02, 0xbb, 0x00, 0xbe, 0x00},
	}

	got, err := ParseFrame(data[0])
	if err != nil {
		t.Error(err)
	}
	log.Printf("[%v] %v\n", "", got)

	want := Frame{
		Data:  Data{0x10, 0x81, 0x0, 0x0, 0x05, 0xff, 0x01, 0x01, 0x30, 0x01, 0x62, 0x02, 0xbb, 0x00, 0xbe, 0x00},
		EHD:   Data{0x10, 0x81},
		TID:   Data{0x00, 0x00},
		EData: Data{0x05, 0xff, 0x01, 0x01, 0x30, 0x01, 0x62, 0x02, 0xbb, 0x00, 0xbe, 0x00},
		SEOJ:  Data{0x05, 0xff, 0x01},
		DEOJ:  Data{0x01, 0x30, 0x01},
		ESV:   ESVType(0x62),
		OPC:   Data{0x02},
		Properties: []Property{
			{Code: 0xbb, Len: 0, Data: []byte{}},
			{Code: 0xbe, Len: 0, Data: []byte{}},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ParseFrame differs: (-want +got)\n%s", diff)
	}
}
