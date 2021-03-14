// +build medium

package wisun

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

const (
	testPort = "COM2"
)

func Test_Medium_Version(t *testing.T) {

	wisunClient := NewBP35C2Client(testPort)
	defer wisunClient.Close()

	got, err := wisunClient.Version()
	if err != nil {
		t.Fatal("failed to exec Version")
	}
	if got != "1.0.0" {
		t.Fatalf("different result %s", got)
	}
}

func Test_Medium_SetBRoutePassword(t *testing.T) {
	c := NewBP35C2Client(testPort)
	defer c.Close()

	err := c.SetBRoutePassword("TESTPWDYYYYY")
	if err != nil {
		t.Fatalf("test failed: %s", err)
	}
}

func Test_Medium_SetBRouteID(t *testing.T) {
	c := NewBP35C2Client(testPort)
	defer c.Close()

	err := c.SetBRouteID("000000TESTID00000000000000000000")
	if err != nil {
		t.Fatalf("test failed: %s", err)
	}
}

func Test_Medium_scan(t *testing.T) {
	testcases := []struct {
		name     string
		duration int
		want     bool
	}{
		{name: "not found", duration: 4, want: false},
		{name: "found", duration: 5, want: true},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			c := NewBP35C2Client(testPort)
			defer c.Close()

			got := c.scan(tc.duration)
			if tc.want != got {
				t.Errorf("Diffrent result: want:%v, got:%v", tc.want, got)
			}
		})
	}
}

func Test_Medium_Scan(t *testing.T) {
	testcases := []struct {
		name string
		want PanDesc
		err  error
	}{
		{
			name: "found",
			want: PanDesc{
				Addr:     "12345678ABCDEF01",
				IPV6Addr: "",
				Channel:  "21",
				PanID:    "8888",
			},
		},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			c := NewBP35C2Client(testPort)
			defer c.Close()

			got, err := c.Scan()
			if tc.want != got {
				t.Errorf("Diffrent result: want:%v, got:%v", tc.want, got)
			}
			if tc.err != err {
				t.Errorf("Diffrent err: want:%v, got:%v", tc.err, err)
			}

		})
	}

}

func Test_Medium_LL64(t *testing.T) {
	c := NewBP35C2Client(testPort)
	defer c.Close()

	want := "FE80:0000:0000:0000:021D:1290:1234:ABCD"

	got, err := c.LL64("12345678ABCDEF01")
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Diffrent result: -want, +got: \n%s", diff)
	}

	if err != nil {
		t.Errorf("%s", err)
	}
}

func Test_Medium_SRegS2(t *testing.T) {

	testcases := []struct {
		name    string
		channel string
		err     error
	}{
		{
			name:    "success",
			channel: "21",
			err:     nil,
		},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			c := NewBP35C2Client(testPort)
			defer c.Close()

			err := c.SRegS2(tc.channel)
			if tc.err != err {
				t.Errorf("Diffrent result: want:%v, got:%v", tc.err, err)
			}
		})
	}
}

func Test_Medium_SRegS3(t *testing.T) {

	testcases := []struct {
		name  string
		panID string
		err   error
	}{
		{name: "success", panID: "0002", err: nil},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			c := NewBP35C2Client(testPort)
			defer c.Close()

			err := c.SRegS3(tc.panID)
			if tc.err != err {
				t.Errorf("Diffrent result: want:%v, got:%v", tc.err, err)
			}
		})
	}
}

func Test_Medium_Join(t *testing.T) {
	testcases := []struct {
		name    string
		panDesc PanDesc
		want    bool
		err     error
	}{
		{
			name:    "success",
			panDesc: PanDesc{},
			want:    true,
			err:     nil,
		},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			c := NewBP35C2Client(testPort)
			defer c.Close()

			got, err := c.Join(tc.panDesc)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Diffrent result: -want, +got: \n%s", diff)
			}

			if tc.err != err {
				t.Errorf("Diffrent error: want:%v, got:%v", tc.err, err)
			}
		})
	}
}

func Test_Medium_Send(t *testing.T) {

	testcases := []struct {
		name     string
		ipv6addr string
		data     []byte
		want     []byte
		err      error
	}{
		{
			name: "success",
			data: []byte{'X', 'X', 'X', 'X'},
			want: []byte{0x10, 0x81, 0x00, 0x01, 0x02, 0x88, 0x01, 0x05, 0xff, 0x01, 'r', 0x01, 0xe7, 0x04, 0x00, 0x00, 0x01, 0xf8},
			err:  nil,
		},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			c := NewBP35C2Client(testPort)
			defer c.Close()

			got, err := c.Send(tc.data)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Diffrent result: -want, +got: \n%s", diff)
			}

			if tc.err != err {
				t.Errorf("Diffrent error: want:%v, got:%v", tc.err, err)
			}
		})
	}
}
