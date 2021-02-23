package main

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/u-one/go-el-controller/transport"
)

func TestController(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	r := transport.NewMockMulticastReceiver(ctrl)
	ur := transport.NewMockUnicastReceiver(ctrl)
	s := transport.NewMockMulticastSender(ctrl)

	mch := make(chan transport.ReceiveResult, 1)
	defer close(mch)
	uch := make(chan transport.ReceiveResult, 1)
	defer close(uch)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	r.EXPECT().Start(gomock.Any(), "224.0.23.0", ":3610").Return(mch)
	ur.EXPECT().Start(gomock.Any(), ":3610").Return(uch)

	c := ControllerNode{
		MulticastReceiver: r,
		MulticastSender:   s,
		UnicastReceiver:   ur,
	}

	state := 0

	infFrame := []byte{0x10, 0x81, 0x0, 0x0, 0x0e, 0xf0, 0x01, 0x0e, 0xf0, 0x01, 0x73, 0x01, 0xd5, 0x04, 0x01, 0x05, 0xff, 0x01}
	s.EXPECT().Send(infFrame).Do(func(data []byte) {
		fmt.Println("infFrame sent", state)
		if state != 0 {
			t.Errorf("order not correct")
		}
		state++
	})

	infReqFrame := []byte{0x10, 0x81, 0x0, 0x1, 0x05, 0xff, 0x01, 0x0e, 0xf0, 0x01, 0x63, 0x01, 0xd5, 0x00}
	s.EXPECT().Send(infReqFrame).Do(func(data []byte) {
		fmt.Println("infReqFrame sent", state)
		if state != 1 {
			t.Errorf("order not correct")
		}
		state++
	})

	getFrame := []byte{0x10, 0x81, 0x0, 0x2, 0x05, 0xff, 0x01, 0x0e, 0xf0, 0x01, 0x62, 0x08, 0x80, 0x00, 0x82, 0x00, 0xd3, 0x00, 0xd4, 0x00, 0xd5, 0x00, 0xd6, 0x00, 0xd7, 0x00, 0x9f, 0x00}
	s.EXPECT().Send(getFrame).Do(func(data []byte) {
		fmt.Println("getFrame sent", state)

		if state != 2 {
			t.Errorf("order not correct")
		}
		state++
	})

	s.EXPECT().Send(gomock.Any()).Do(func(indata []byte) {
		fmt.Println("aircongetFrame sent", state)

		aircongetFrame := []byte{0x10, 0x81, 0x0, byte(state), 0x05, 0xff, 0x01, 0x01, 0x30, 0x01, 0x62, 0x04, 0x81, 0x00, 0x83, 0x00, 0xbb, 0x00, 0xbe, 0x00}
		if !bytes.Equal(aircongetFrame, indata) {
			t.Errorf("input is not correct")
		}

		if state < 3 {
			t.Errorf("order not correct")
		}
		state++

		send := func(data []byte, addr string, delay time.Duration) {
			timer := time.NewTimer(delay)
			<-timer.C
			rr := transport.ReceiveResult{Data: data, Address: addr, Err: nil}
			mch <- rr
		}
		airconResp := []byte{0x10, 0x81, 0x0, 0x0, 0x4, 0x30, 0x1, 0x5, 0xff, 0x1, 0x72, 0x4, 0x81, 0x1, 0x41, 0x83, 0x11, 0xfe, 0x0, 0x0, 0x8, 0x60, 0xf1, 0x89, 0x30, 0x6d, 0xf5, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xbb, 0x01, 0x01, 0xbe, 0x1, 0x09}
		send(airconResp, "192.168.1.1", time.Second)
		send(airconResp, "192.168.1.2", time.Millisecond*200)

		if state > 3 {
			timer := time.NewTimer(time.Second)
			<-timer.C
			cancel()
		}
	}).AnyTimes()

	c.Start(ctx)

	<-ctx.Done()

}
