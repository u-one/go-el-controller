package echonetlite

import (
	"encoding/hex"
	"log"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func init() {
	log.SetOutput(os.Stdout)
}

func toData(t *testing.T, s string) Data {
	t.Helper()

	data, err := hex.DecodeString(s)
	if err != nil {
		t.Error(err)
	}
	return Data(data)
}

func toByteArray(t *testing.T, s string) []byte {
	t.Helper()

	b, err := hex.DecodeString(s)
	if err != nil {
		t.Error(err)
	}
	return b
}

//[192.168.1.10:40929] 108100390ef0010ef0016201d600
//[192.168.1.10:49322] 108100410ef0010ef0017301d5040105ff01
//[192.168.1.17:4527] 108100050130010ef0017301800130
//[192.168.1.17:4530] 108100060130010ef0017301800131

func Test_ParseFrame(t *testing.T) {

	testcases := []struct {
		name      string
		input     []byte
		wantFrame Frame
		wantData  Data
		wantEData Data
	}{
		{
			name:  "test1",
			input: []byte{0x10, 0x81, 0x0, 0x0, 0x05, 0xff, 0x01, 0x01, 0x30, 0x01, 0x62, 0x02, 0xbb, 0x00, 0xbe, 0x00},
			wantFrame: Frame{
				EHD:  Data{0x10, 0x81},
				TID:  Data{0x00, 0x00},
				SEOJ: Object{Data{0x05, 0xff, 0x01}},
				DEOJ: Object{Data{0x01, 0x30, 0x01}},
				ESV:  ESVType(0x62),
				OPC:  0x02,
				Properties: []Property{
					{Code: 0xbb, Len: 0, Data: []byte{}},
					{Code: 0xbe, Len: 0, Data: []byte{}},
				},
			},
			wantData:  Data{0x10, 0x81, 0x0, 0x0, 0x05, 0xff, 0x01, 0x01, 0x30, 0x01, 0x62, 0x02, 0xbb, 0x00, 0xbe, 0x00},
			wantEData: Data{0x05, 0xff, 0x01, 0x01, 0x30, 0x01, 0x62, 0x02, 0xbb, 0x00, 0xbe, 0x00},
		},
		{
			name: "test2",
			// []byteのリテラルだと長くなるのでstringから変換できるパターンも用意してみた
			input: toByteArray(t, "1081000001300105ff0172048101418311fe00000860f189306df500000000000000bb011cbe0119"),
			wantFrame: Frame{
				EHD:  toData(t, "1081"),
				TID:  toData(t, "0000"),
				SEOJ: Object{toData(t, "013001")},
				DEOJ: Object{toData(t, "05ff01")},
				ESV:  GetRes,
				OPC:  0x04,
				Properties: []Property{
					{Code: 0x81, Len: 1, Data: toData(t, "41")},
					{Code: 0x83, Len: 17, Data: toData(t, "fe00000860f189306df500000000000000")},
					{Code: 0xbb, Len: 1, Data: toData(t, "1c")},
					{Code: 0xbe, Len: 1, Data: toData(t, "19")},
				},
				Object: AirconObject{
					SuperObject:  SuperObject{InstallLocation: Location{Code: Room, Number: 1}},
					InternalTemp: 28,
					OuterTemp:    25,
				},
			},
			wantData:  toData(t, "1081000001300105ff0172048101418311fe00000860f189306df500000000000000bb011cbe0119"),
			wantEData: toData(t, "01300105ff0172048101418311fe00000860f189306df500000000000000bb011cbe0119"),
		},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseFrame(tc.input)
			if err != nil {
				t.Error(err)
			}
			log.Printf("[%v] %v\n", "", got)

			if diff := cmp.Diff(tc.wantFrame, got); diff != "" {
				t.Errorf("ParseFrame differs: (-want +got)\n%s", diff)
			}

			gotEData := got.EData()
			if diff := cmp.Diff(tc.wantEData, gotEData); diff != "" {
				t.Errorf("ParseFrame differs: (-want +got)\n%s", diff)
			}

			gotData := got.Serialize()
			if diff := cmp.Diff(tc.wantData, gotData); diff != "" {
				t.Errorf("ParseFrame differs: (-want +got)\n%s", diff)
			}
		})
	}

}

func TestCreateAirconGetFrame(t *testing.T) {

	got := CreateAirconGetFrame(0x0)

	want := &Frame{
		EHD:  Data{0x10, 0x81},
		TID:  Data{0x00, 0x00},
		SEOJ: Object{Data{0x05, 0xff, 0x01}},
		DEOJ: Object{Data{0x01, 0x30, 0x01}},
		ESV:  ESVType(0x62),
		OPC:  0x04,
		Properties: []Property{
			{Code: 0x81, Len: 0, Data: []byte{}},
			{Code: 0x83, Len: 0, Data: []byte{}},
			{Code: 0xbb, Len: 0, Data: []byte{}},
			{Code: 0xbe, Len: 0, Data: []byte{}},
		},
	}

	wantData := Data{0x10, 0x81, 0x0, 0x0, 0x05, 0xff, 0x01, 0x01, 0x30, 0x01, 0x62, 0x04, 0x81, 0x00, 0x83, 0x00, 0xbb, 0x00, 0xbe, 0x00}
	wantEData := Data{0x05, 0xff, 0x01, 0x01, 0x30, 0x01, 0x62, 0x04, 0x81, 0x00, 0x83, 0x00, 0xbb, 0x00, 0xbe, 0x00}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ParseFrame differs: (-want +got)\n%s", diff)
	}

	gotEData := want.EData()
	if diff := cmp.Diff(wantEData, gotEData); diff != "" {
		t.Errorf("ParseFrame differs: (-want +got)\n%s", diff)
	}

	gotData := want.Serialize()
	if diff := cmp.Diff(wantData, gotData); diff != "" {
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

func TestNewFrame(t *testing.T) {
	props := []Property{}
	props = append(props, Property{Code: 0xd5, Len: 4, Data: []byte{0x01, 0x05, 0xff, 0x01}})

	f := NewFrame(1, props)

	got := f.Serialize()
	want := Data{0x10, 0x81, 0x00, 0x01, 0x0e, 0xf0, 0x01, 0x0e, 0xf0, 0x01, 0x73, 0x01, 0xd5, 0x04, 0x01, 0x05, 0xff, 0x01}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}

}

func TestCreateInfFrame(t *testing.T) {

	got := CreateInfFrame(1)

	wantdata := []byte{0x10, 0x81, 0x00, 0x01, 0x0e, 0xf0, 0x01, 0x0e, 0xf0, 0x01, 0x73, 0x01, 0xd5, 0x04, 0x01, 0x05, 0xff, 0x01}

	want, err := ParseFrame(wantdata)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if diff := cmp.Diff(want, *got); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}

}
