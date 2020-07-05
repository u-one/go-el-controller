package wisun

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
)

func TestVer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := NewMockSerialClient(ctrl)
	mock.EXPECT().Send(gomock.Any()).Return(nil)
	mock.EXPECT().Recv().Return("OK", nil).Times(3)

	c := NewElectricityMeterClient(mock)
	c.Version()

}

func TestStartSequence(t *testing.T) {

	pattern := map[string][]string{
		"SKVER\r\n": []string{
			"EVER 1.5.2\r\n",
			"OK\r\n",
		},
		"SKSETRBID 000000TESTID00000000000000000000\r\n": []string{
			"OK\r\n",
		},
		"SKSETPWD C TESTPWDYYYYY\r\n": []string{
			"OK\r\n",
		},
		"SKSCAN 2 FFFFFFFF 4 0 \r\n": []string{
			"OK\r\n",
			"EVENT 22 2001:0DB8:0000:0000:011A:1111:0000:0001 0\r\n",
		},
		"SKSCAN 2 FFFFFFFF 5 0 \r\n": []string{
			"OK\r\n",
			"EVENT 20 2001:0DB8:0000:0000:011A:1111:0000:0001 0\r\n",
			"EPANDESC\r\n",
			"  Channel:21\r\n",
			"  Channel Page:01\r\n",
			"  Pan ID:0002\r\n",
			"  Addr:001A111100000002\r\n",
			"  LQI:CA\r\n",
			"  Side:0\r\n",
			"  PairID:0112CE67\r\n",
		},
		"SKLL64 001A111100000002\r\n": []string{
			"2001:0DB8:0000:0000:011A:1111:0000:0002\r\n",
		},
		"SKSREG S2 21\r\n": []string{
			"OK\r\n",
		},
		"SKSREG S3 0002\r\n": []string{
			"OK\r\n",
		},
		"SKJOIN 2001:0DB8:0000:0000:011A:1111:0000:0002\r\n": []string{
			"OK\r\n",
			"EVENT 21 2001:0DB8:0000:0000:011A:1111:0000:0002 0 00\r\n",
			"EVENT 21 2001:0DB8:0000:0000:011A:1111:0000:0002 0 00\r\n",
			"ERXUDP 21 2001:0DB8:0000:0000:011A:1111:0000:0002 2001:0DB8:0000:0000:011A:1111:0000:0001 02CC 02CC 001C6400030C12A4 0 0 0068 \x00\x00\x00h\x80\x00\x00\x02\x1c/\xf4\xb1\xd1\xa2\x95\xa9\x00\x05\x00\x00\x00\x10\x00\x00KtCV\xb5^\xdd&\x17$)\xe8;\xc67*\x00\x02\x00\x00\x008\x00\x00\x01Z\x008/\x00\x8a\x7f\x84+\xd4\x1bL\t\x02%\x8ey\x1a\x8f\xf6\x05SM0000009902110000000000000112CE67\r\n",
			"EVENT 21 2001:0DB8:0000:0000:011A:1111:0000:0002 0 00\r\n",
			"ERXUDP 2001:0DB8:0000:0000:011A:1111:0000:0002 2001:0DB8:0000:0000:011A:1111:0000:0001 02CC 02CC 001C6400030C12A4 0 0 0054 \x00\x00\x00T\x80\x00\x00\x02\x1c/\xf4\xb1\xd1\xa2\x95\xaa\x00\x02\x00\x00\x00;\x00\x00\x01[\x00;/\x80\x8a\x7f\x84+\xd4\x1bL\t\x02%\x8ey\x1a\x8f\xf6\x05\x91M\xa6\xe5\x0c\x90\xed\xe8\xac\xc0^\x03Yy\xbaJ\x00\x00\x00\x00\xc69\xce\r\x16\xcct4\x9c\x8fm\xf9\xff\x9dn\xa1\xd2\x00\r\n",
			"EVENT 21 2001:0DB8:0000:0000:011A:1111:0000:0002 0 00\r\n",
			"ERXUDP 2001:0DB8:0000:0000:011A:1111:0000:0002 2001:0DB8:0000:0000:011A:1111:0000:0001 02CC 02CC 001C6400030C12A4 0 0 0058 \x00\x00\x00X\xa0\x00\x00\x02\x1c/\xf4\xb1\xd1\xa2\x95\xab\x00\x07\x00\x00\x00\x04\x00\x00\x00\x00\x00\x00\x00\x02\x00\x00\x00\x04\x00\x00\x03[\x00\x04\x00\x04\x00\x00\x00\x04\x00\x00\x00\x00\t\x01\x00\x08\x00\x00\x00\x04\x00\x00\x00\x01Q\x80\x00\x01\x00\x00\x00\x10\x00\x00\xb9\xb1\xf3\xab+\xe9\xc5y\x0e\xbf\xb6\x14t\xb3J\xaf\r\n",
			"EVENT 21 2001:0DB8:0000:0000:011A:1111:0000:0002 0 00\r\n",
			"EVENT 25 2001:0DB8:0000:0000:011A:1111:0000:0002 0\r\n",
			"ERXUDP 2001:0DB8:0000:0000:011A:1111:0000:0002 2001:0DB8:0000:0000:011A:1111:0000:0001 02CC 02CC 001A111100000002 0 0 0028 \x10\x81\x00\x00\x0e\xf0\x01\x0e\xf0\x01s\x01\xd5\x04\x01\x02\x88\x01\r\n", // TODO: put binary
			"\r\n",
			"\r\n",
			"\r\n",
			"\r\n",
		},
		"SKSENDTO 1 2001:0DB8:0000:0000:011A:1111:0000:0002 0E1A 1 0 000E \r\n": []string{
			"EVENT 21 2001:0DB8:0000:0000:011A:1111:0000:0002 0 00\r\n",
			"OK\r\n",
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := NewMockSerialClient(ctrl)

	lastCmd := ""
	respCnt := -1

	mock.EXPECT().Send(gomock.Any()).DoAndReturn(func(cmd string) error {
		lastCmd = cmd
		respCnt = -1
		return nil
	}).AnyTimes()

	mock.EXPECT().Recv().DoAndReturn(func() (string, error) {
		responses := pattern[lastCmd]
		resp := ""
		if respCnt == -1 {
			resp = lastCmd
		} else {
			resp = responses[respCnt]
		}
		respCnt++
		return resp, nil
	}).AnyTimes()

	c := NewElectricityMeterClient(mock)
	c.StartSequence("000000TESTID00000000000000000000", "TESTPWDYYYYY")

	/*
		for _, c := range comms {
			mock.EXPECT().Send(c.cmd).Return(c.res, nil).AnyTimes()
		}

		for _, tc := range testcases {
			_, err := mock.Send(tc.cmd)
			if err != nil {
				t.Fatalf("failed")
			}
		}
	*/

}
