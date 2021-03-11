package wisun

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/u-one/go-el-controller/transport"
)

type resp struct {
	d string
	e error
}

func mock(t *testing.T, m *transport.MockSerial, input string, response []resp) {
	t.Helper()

	lastCmd := ""
	respCnt := -1

	m.EXPECT().Send(gomock.Any()).DoAndReturn(func(cmd []byte) error {
		lastCmd = string(cmd)
		respCnt = -1
		return nil
	}).AnyTimes()

	m.EXPECT().Recv().DoAndReturn(func() ([]byte, error) {
		resp := ""
		var err error
		if respCnt == -1 {
			resp = lastCmd
		} else {
			resp = response[respCnt].d
			err = response[respCnt].e
		}
		respCnt++
		return []byte(resp), err
	}).AnyTimes()

}

func Test_Version(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := transport.NewMockSerial(ctrl)
	mock(t, m, "SKVER\r\n", []resp{
		{"EVER 1.5.2\r\n", nil},
		{"OK\r\n", nil},
	})

	c := &BP35C2Client{serial: m}
	c.Version()
}

func Test_SetBRoutePassword(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := transport.NewMockSerial(ctrl)
	mock(t, m, "SKSETPWD C TESTPWDYYYYY\r\n", []resp{
		{"OK\r\n", nil},
	})

	c := &BP35C2Client{serial: m}
	c.SetBRoutePassword("TESTPWDYYYYY")
}

func Test_SetBRouteID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := transport.NewMockSerial(ctrl)
	mock(t, m, "SKSETRBID 000000TESTID00000000000000000000\r\n", []resp{
		{"OK\r\n", nil},
	})

	c := &BP35C2Client{serial: m}
	c.SetBRouteID("000000TESTID00000000000000000000")
}

func Test_scan(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name     string
		duration int
		input    string
		response []resp
		expect   bool
	}{
		{
			name:     "not found",
			duration: 4,
			input:    "SKSCAN 2 FFFFFFFF 4 0 \r\n",
			response: []resp{
				{"OK\r\n", nil},
				{"EVENT 22 2001:0DB8:0000:0000:011A:1111:0000:0001 0\r\n", nil},
			},
			expect: false,
		},
		{
			name:     "found",
			duration: 5,
			input:    "SKSCAN 2 FFFFFFFF 5 0 \r\n",
			response: []resp{
				{"OK\r\n", nil},
				{"EVENT 20 2001:0DB8:0000:0000:011A:1111:0000:0001 0\r\n", nil},
			},
			expect: true,
		},

		/* TODO: fix
		   {
		   	name:     "fail",
		   	duration: 5,
		   	input:    "SKSCAN 2 FFFFFFFF 5 0 \r\n",
		   	response: []resp{
		   		{"FAIL\r\n", nil},
		   	},
		   	expect: false,
		   },
		   {
		   	name:     "fail",
		   	duration: 5,
		   	input:    "SKSCAN 2 FFFFFFFF 5 0 \r\n",
		   	response: []resp{
		   		{"\r\n", fmt.Errorf("error")},
		   	},
		   	expect: false,
		   },
		*/

	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := transport.NewMockSerial(ctrl)
			mock(t, m, tc.input, tc.response)

			c := &BP35C2Client{serial: m}
			got := c.scan(tc.duration)
			if tc.expect != got {
				t.Errorf("Diffrent result: want:%v, got:%v", tc.expect, got)
			}
		})
	}
}

func Test_receivePanDesc(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := transport.NewMockSerial(ctrl)

	response := []string{
		"EPANDESC\r\n",
		"  Channel:21\r\n",
		"  Channel Page:01\r\n",
		"  Pan ID:0002\r\n",
		"  Addr:001A111100000002\r\n",
		"  LQI:CA\r\n",
		"  Side:0\r\n",
		"  PairID:0112CE67\r\n",
	}

	respCnt := 0
	m.EXPECT().Recv().DoAndReturn(func() ([]byte, error) {
		resp := response[respCnt]
		respCnt++
		return []byte(resp), nil
	}).AnyTimes()

	want := PanDesc{
		Addr:     "001A111100000002",
		IPV6Addr: "",
		Channel:  "21",
		PanID:    "0002",
	}

	c := &BP35C2Client{serial: m}
	got, err := c.receivePanDesc()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Diffrent result: -want, +got: \n%s", diff)
	}

	if err != nil {
		t.Errorf("%s", err)
	}
}

func Test_Scan(t *testing.T) {
	t.Parallel()
	// TODO: implement
}

func Test_LL64(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := transport.NewMockSerial(ctrl)
	mock(t, m, "001A111100000002", []resp{
		{"2001:0DB8:0000:0000:011A:1111:0000:0002\r\n", nil},
	})

	want := "2001:0DB8:0000:0000:011A:1111:0000:0002"

	c := &BP35C2Client{serial: m}
	got, err := c.LL64("001A111100000002")
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Diffrent result: -want, +got: \n%s", diff)
	}

	if err != nil {
		t.Errorf("%s", err)
	}
}

func Test_SRegS2(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name     string
		channel  string
		input    string
		response []resp
		err      error
	}{
		{
			name:    "success",
			channel: "21",
			input:   "SKSREG S2 21\r\n",
			response: []resp{
				{"OK\r\n", nil},
			},
			err: nil,
		},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := transport.NewMockSerial(ctrl)
			mock(t, m, tc.input, tc.response)

			c := &BP35C2Client{serial: m}
			err := c.SRegS2(tc.channel)
			if tc.err != err {
				t.Errorf("Diffrent result: want:%v, got:%v", tc.err, err)
			}
		})
	}
}

func Test_SRegS3(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name     string
		panID    string
		input    string
		response []resp
		err      error
	}{
		{
			name:  "success",
			panID: "0002",
			input: "SKSREG S3 0002\r\n",
			response: []resp{
				{"OK\r\n", nil},
			},
			err: nil,
		},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := transport.NewMockSerial(ctrl)
			mock(t, m, tc.input, tc.response)

			c := &BP35C2Client{serial: m}
			err := c.SRegS3(tc.panID)
			if tc.err != err {
				t.Errorf("Diffrent result: want:%v, got:%v", tc.err, err)
			}
		})
	}
}

func Test_Join(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name     string
		panDesc  PanDesc
		input    string
		response []resp
		want     bool
		err      error
	}{
		{
			name:    "success",
			panDesc: PanDesc{},
			input:   "SKJOIN 2001:0DB8:0000:0000:011A:1111:0000:0002\r\n",
			response: []resp{
				{"OK\r\n", nil},
				{"EVENT 21 2001:0DB8:0000:0000:011A:1111:0000:0002 0 00\r\n", nil},
				{"ERXUDP 2001:0DB8:0000:0000:011A:1111:0000:0002 2001:0DB8:0000:0000:011A:1111:0000:0001 02CC 02CC 001C6400030C12A4 0 0 0054 \x00\x00\x00T\x80\x00\x00\x02\x1c/\xf4\xb1\xd1\xa2\x95\xaa\x00\x02\x00\x00\x00;\x00\x00\x01[\x00;/\x80\x8a\x7f\x84+\xd4\x1bL\t\x02%\x8ey\x1a\x8f\xf6\x05\x91M\xa6\xe5\x0c\x90\xed\xe8\xac\xc0^\x03Yy\xbaJ\x00\x00\x00\x00\xc69\xce\r\x16\xcct4\x9c\x8fm\xf9\xff\x9dn\xa1\xd2\x00\r\n", nil},
				{"EVENT 25 2001:0DB8:0000:0000:011A:1111:0000:0002 0\r\n", nil},
				{"ERXUDP 2001:0DB8:0000:0000:011A:1111:0000:0002 2001:0DB8:0000:0000:011A:1111:0000:0001 02CC 02CC 001A111100000002 0 0 0028 \x10\x81\x00\x00\x0e\xf0\x01\x0e\xf0\x01s\x01\xd5\x04\x01\x02\x88\x01\r\n", nil},
				{"\r\n", nil},
				{"\r\n", nil},
				{"\r\n", nil},
				{"\r\n", nil},
			},
			want: true,
			err:  nil,
		},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := transport.NewMockSerial(ctrl)
			mock(t, m, tc.input, tc.response)

			c := &BP35C2Client{serial: m}
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

func Test_Send(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name     string
		ipv6addr string
		data     []byte
		input    string
		response []resp
		want     []byte
		err      error
	}{
		{
			name:  "success",
			data:  []byte{'X', 'X', 'X', 'X'},
			input: "SKSENDTO 1 2001:0DB8:0000:0000:011A:1111:0000:0002 0E1A 1 0 000E \r\n",
			response: []resp{
				{"EVENT 21 2001:0DB8:0000:0000:011A:1111:0000:0002 0 00\r\n", nil},
				{"OK\r\n", nil},
				{"ERXUDP FE80:0000:0000:0000:021C:6400:030C:12A4 FE80:0000:0000:0000:021D:1291:0000:0574 0E1A 0E1A 001C6400030C12A4 1 0 0012 \x10\x81\x00\x01\x02\x88\x01\x05\xff\x01r\x01\xe7\x04\x00\x00\x01\xf8\r\n", nil},
			},
			want: []byte{0x10, 0x81, 0x00, 0x01, 0x02, 0x88, 0x01, 0x05, 0xff, 0x01, 'r', 0x01, 0xe7, 0x04, 0x00, 0x00, 0x01, 0xf8},
			err:  nil,
		},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := transport.NewMockSerial(ctrl)
			mock(t, m, tc.input, tc.response)

			c := &BP35C2Client{serial: m, panDesc: PanDesc{IPV6Addr: "2001:0DB8:0000:0000:011A:1111:0000:0002"}}
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

func Test_parseRXUDP(t *testing.T) {
	t.Parallel()

	line := []byte("ERXUDP FE80:0000:0000:0000:021C:6400:030C:12A4 FE80:0000:0000:0000:021D:1291:0000:0574 0E1A 0E1A 001C6400030C12A4 1 0 0012 \x10\x81\x00\x01\x02\x88\x01\x05\xff\x01\x01\xe7\x04\x00\x00\x01\xf8")
	got, err := parseRXUDP(line)
	if err != nil {
		t.Fatalf("error occured")
	}
	want := []byte{0x10, 0x81, 0x00, 0x01, 0x02, 0x88, 0x01, 0x05, 0xff, 0x01, 0x01, 0xe7, 0x04, 0x00, 0x00, 0x01, 0xf8}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("data differs: (-want +got)\n%s", diff)
	}
}
