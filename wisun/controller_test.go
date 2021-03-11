package wisun

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
)

func TestControllerVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := NewMockClient(ctrl)
	mock.EXPECT().Version().Return(nil)

	c := NewController(mock)
	err := c.Version()
	if err != nil {
		t.Fatalf("error occured:%s", err)
	}
}

func TestControllerConnect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := NewMockClient(ctrl)

	mock.EXPECT().SetBRoutePassword("TESTPWDYYYYY").Return(nil)
	mock.EXPECT().SetBRouteID("000000TESTID00000000000000000000").Return(nil)
	mock.EXPECT().Scan().Return(PanDesc{
		Addr:     "001A111100000002",
		IPV6Addr: "",
		Channel:  "21",
		PanID:    "0002",
	}, nil)

	mock.EXPECT().LL64("001A111100000002").Return("2001:0DB8:0000:0000:011A:1111:0000:0002", nil)
	mock.EXPECT().SRegS2("21").Return(nil)
	mock.EXPECT().SRegS3("0002").Return(nil)
	mock.EXPECT().Join(PanDesc{
		Addr:     "001A111100000002",
		IPV6Addr: "2001:0DB8:0000:0000:011A:1111:0000:0002",
		Channel:  "21",
		PanID:    "0002",
	}).Return(true, nil)

	c := NewController(mock)
	err := c.Connect("000000TESTID00000000000000000000", "TESTPWDYYYYY")
	if err != nil {
		t.Errorf("failed: %s", err)
	}
}

func TestControllerGetCurrentConsumption(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := NewMockClient(ctrl)

	mock.EXPECT().SendTo(
		"FE80:0000:0000:0000:021C:6400:030C:12A4",
		[]byte{0x10, 0x81, 0x00, 0x01, 0x05, 0xff, 0x01, 0x02, 0x88, 0x01, 0x62, 0x01, 0xe7, 0x00},
	).Return(
		[]byte{0x10, 0x81, 0x00, 0x01, 0x02, 0x88, 0x01, 0x05, 0xff, 0x01, 0x72, 0x01, 0xe7, 0x04, 0x00, 0x00, 0x01, 0xf8}, nil,
	)

	c := Controller{client: mock, panDesc: PanDesc{IPV6Addr: "FE80:0000:0000:0000:021C:6400:030C:12A4"}}
	p, err := c.GetCurrentPowerConsumption()
	if err != nil {
		t.Fatalf("error occured:%s", err)
	}

	if p != 0 {
		t.Errorf("different")
	}
}
