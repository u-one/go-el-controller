package wisun

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
)

func TestCommand(t *testing.T) {

	comms := []struct {
		cmd string
		res []string
	}{
		{
			cmd: "SKVER\r\n",
			res: []string{
				"EVER 1.5.2\r\n",
				"OK\r\n",
			},
		},
		{
			cmd: "SKSETRBID 0000009902110000000000000112CE67\r\n",
			res: []string{
				"OK\r\n",
			},
		},
		{
			cmd: "SKSETPWD C 4DWNPK18RR5V\r\n",
			res: []string{
				"OK\r\n",
			},
		},
		{
			cmd: "SKSCAN 2 FFFFFFFF 4 0 \r\n",
			res: []string{
				"OK\r\n",
				"EVENT 22 FE80:0000:0000:0000:021D:1291:0000:0574 0\r\n",
			},
		},
		{
			cmd: "SKSCAN 2 FFFFFFFF 5 0 \r\n",
			res: []string{
				"OK\r\n",
				"EVENT 20 FE80:0000:0000:0000:021D:1291:0000:0574 0\r\n",
				"EPANDESC\r\n",
				"  Channel:33\r\n",
				"  Channel Page:09\r\n",
				"  Pan ID:12A4\r\n",
				"  Addr:001C6400030C12A4\r\n",
				"  LQI:CA\r\n",
				"  Side:0\r\n",
				"  PairID:0112CE67\r\n",
			},
		},
	}

	testcases := []struct {
		cmd string
	}{
		{cmd: "SKVER\r\n"},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := NewMockSerial(ctrl)
	for _, c := range comms {
		mock.EXPECT().Send(c.cmd).Return(c.res, nil).AnyTimes()
	}

	for _, tc := range testcases {
		_, err := mock.Send(tc.cmd)
		if err != nil {
			t.Fatalf("failed")
		}
	}

}
