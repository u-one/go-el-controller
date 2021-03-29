package echonetlite

import (
	"context"
	"fmt"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/u-one/go-el-controller/wisun"
)

func TestStart(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testcases := []struct {
		name   string
		brID   string
		brPW   string
		client func(*wisun.MockClient)
		err    error
	}{
		{
			name: "success",
			brID: "0123456789AB",
			brPW: "00112233445566778899AABBCCDDEEFF",
			client: func(m *wisun.MockClient) {
				m.EXPECT().Connect(ctx, "0123456789AB", "00112233445566778899AABBCCDDEEFF").Return(nil)
			},
			err: nil,
		},
		{
			name: "failure",
			brID: "0123456789AB",
			brPW: "00112233445566778899AABBCCDDEEFF",
			client: func(m *wisun.MockClient) {
				m.EXPECT().Connect(ctx, "0123456789AB", "00112233445566778899AABBCCDDEEFF").Return(fmt.Errorf("error"))
			},
			err: fmt.Errorf("exec Connect failed: error"),
		},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mock := wisun.NewMockClient(ctrl)
			tc.client(mock)

			node := NewElectricityControllerNode(mock)
			err := node.Start(ctx, tc.brID, tc.brPW)

			if tc.err != nil && err != nil {
				if tc.err.Error() != err.Error() {
					t.Errorf("Diffrent result: want:%#v, got:%#v", tc.err, err)
				}
			} else if tc.err != err {
				t.Errorf("Diffrent result: want:%#v, got:%#v", tc.err, err)
			}

		})
	}
}

func TestClose(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := wisun.NewMockClient(ctrl)
	mock.EXPECT().Close()
	node := NewElectricityControllerNode(mock)
	node.Close()
}

func TestGetPowerConsumption(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name   string
		client func(*wisun.MockClient)
		want   int
		err    error
	}{
		{
			name: "success",
			client: func(m *wisun.MockClient) {
				m.EXPECT().
					Send([]byte("\x10\x81\x00\x01\x05\xff\x01\x02\x88\x01\x62\x01\xe7\x00")).
					Return([]byte("\x10\x81\x00\x01\x02\x88\x01\x05\xff\x01\x72\x01\xe7\x04\x00\x00\x01\xf8"), nil)
			},
			want: 504,
			err:  nil,
		},
		{
			name: "failure",
			client: func(m *wisun.MockClient) {
				m.EXPECT().
					Send([]byte("\x10\x81\x00\x01\x05\xff\x01\x02\x88\x01\x62\x01\xe7\x00")).
					Return([]byte{}, fmt.Errorf("error"))
			},
			want: 0,
			err:  fmt.Errorf("error"),
		},
		{
			name: "invalid frame",
			client: func(m *wisun.MockClient) {
				m.EXPECT().
					Send([]byte("\x10\x81\x00\x01\x05\xff\x01\x02\x88\x01\x62\x01\xe7\x00")).
					Return([]byte("\x10\x81"), nil)
			},
			want: 0,
			err:  fmt.Errorf("invalid frame: size is too short:2"),
		},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mock := wisun.NewMockClient(ctrl)
			tc.client(mock)

			node := NewElectricityControllerNode(mock)
			got, err := node.GetPowerConsumption()

			if tc.err != nil && err != nil {
				if tc.err.Error() != err.Error() {
					t.Errorf("Diffrent result: want:%#v, got:%#v", tc.err, err)
				}
			} else if tc.err != err {
				t.Errorf("Diffrent result: want:%#v, got:%#v", tc.err, err)
			}

			if tc.want != got {
				t.Errorf("Diffrent result: want:%#v, got:%#v", tc.want, got)
			}

		})
	}

}

/*
func Test_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := transport.NewMockSerial(ctrl)

	mock(t, m, "SKSENDTO 1 FE80:0000:0000:0000:021C:6400:030C:12A4 0E1A 1 0 000E \x10\x81\x00\x01\x05\xff\x01\x02\x88\x01\x62\x01\xe7\x00\r\n", []resp{
		{"EVENT 21 2001:0DB8:0000:0000:011A:1111:0000:0002 0 00\r\n", nil},
		{"OK\r\n", nil},
		{"ERXUDP FE80:0000:0000:0000:021C:6400:030C:12A4 FE80:0000:0000:0000:021D:1291:0000:0574 0E1A 0E1A 001C6400030C12A4 1 0 0012 \x10\x81\x00\x01\x02\x88\x01\x05\xff\x01\x72\x01\xe7\x04\x00\x00\x01\xf8\r\n", nil},
	})

	c := &BP35C2Client{serial: m, panDesc: PanDesc{IPV6Addr: "FE80:0000:0000:0000:021C:6400:030C:12A4"}}

	p, err := c.Get()
	if err != nil {
		t.Fatalf("error occured:%s", err)
	}

	if p != 0 {
		t.Errorf("different")
	}
}
*/

func TestCreateCurrentPowerConsumptionFrame(t *testing.T) {

	got := CreateCurrentPowerConsumptionFrame(0x0)

	want := &Frame{
		EHD:  Data{EchonetLite, FixedFormat},
		TID:  Data{0x00, 0x00},
		SEOJ: Object{ControllerGroup, Controller, 0x01},
		DEOJ: Object{HomeEquipmentGroup, LowVoltageSmartMeter, 0x01},
		ESV:  Get,
		OPC:  0x01,
		Properties: []Property{
			{Code: byte(InstantPower), Len: 0, Data: []byte{}},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ParseFrame differs: (-want +got)\n%s", diff)
	}
}
