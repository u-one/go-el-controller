package main

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCreateAirconGetFrame(t *testing.T) {

	got := createAirconGetFrame()

	want := &Frame{
		Data:  Data{0x10, 0x81, 0x0, 0x0, 0x05, 0xff, 0x01, 0x01, 0x30, 0x01, 0x62, 0x04, 0x81, 0x00, 0x83, 0x00, 0xbb, 0x00, 0xbe, 0x00},
		EHD:   Data{0x10, 0x81},
		TID:   Data{0x00, 0x00},
		EData: Data{0x05, 0xff, 0x01, 0x01, 0x30, 0x01, 0x62, 0x04, 0x81, 0x00, 0x83, 0x00, 0xbb, 0x00, 0xbe, 0x00},
		SEOJ:  Data{0x05, 0xff, 0x01},
		DEOJ:  Data{0x01, 0x30, 0x01},
		ESV:   ESVType(0x62),
		OPC:   Data{0x04},
		Properties: []Property{
			{Code: 0x81, Len: 0, Data: []byte{}},
			{Code: 0x83, Len: 0, Data: []byte{}},
			{Code: 0xbb, Len: 0, Data: []byte{}},
			{Code: 0xbe, Len: 0, Data: []byte{}},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ParseFrame differs: (-want +got)\n%s", diff)
	}

}

func TestLocationCode(t *testing.T) {

	want := LocationCode(0x1)
	got := Living
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("LocationCode differs: (-want +got)\n%s", diff)
	}

	want = LocationCode(0xF)
	got = Other
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("LocationCode differs: (-want +got)\n%s", diff)
	}
}

func toData(s string) Data {
	data, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return Data(data)
}

func TestParseFrame(t *testing.T) {

	str := "1081000001300105ff0172048101418311fe00000860f189306df500000000000000bb011cbe0119"
	data, err := hex.DecodeString(str)
	fmt.Printf("data: %#v", data)
	if err != nil {
		t.Error(err)
	}

	got, err := ParseFrame(data)
	if err != nil {
		t.Error(err)
	}
	want := Frame{
		Data:  toData("1081000001300105ff0172048101418311fe00000860f189306df500000000000000bb011cbe0119"),
		EHD:   toData("1081"),
		TID:   toData("0000"),
		EData: toData("01300105ff0172048101418311fe00000860f189306df500000000000000bb011cbe0119"),
		SEOJ:  toData("013001"),
		DEOJ:  toData("05ff01"),
		ESV:   GetRes,
		OPC:   toData("04"),
		Properties: []Property{
			{Code: 0x81, Len: 1, Data: toData("41")},
			{Code: 0x83, Len: 17, Data: toData("fe00000860f189306df500000000000000")},
			{Code: 0xbb, Len: 1, Data: toData("1c")},
			{Code: 0xbe, Len: 1, Data: toData("19")},
		},
		Object: AirconObject{
			SuperObject:  SuperObject{InstallLocation: Location{Code: Room, Number: 1}},
			InternalTemp: 28,
			OuterTemp:    25,
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
}