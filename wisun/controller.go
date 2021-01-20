package wisun

import (
	"log"
	"context"
	"fmt"

	"github.com/u-one/go-el-controller/echonetlite"
)

// Controller is
type Controller struct {
	client Client
}

// NewController returns new instance
func NewController(c Client) *Controller {
	return &Controller{c}
}

// Close close
func (c Controller) Close() {
	c.client.Close()
}

// PANAAuth starts PANA authentication
func (c Controller) PANAAuth(bRouteID, bRoutePW string) error {

	if len(bRouteID) == 0 {
		log.Fatal("set B-route ID")
	}
	if len(bRoutePW) == 0 {
		log.Fatal("set B-route password")
	}

	c.client.SetBRoutePassword(bRoutePW)
	c.client.SetBRouteID(bRouteID)

	pd, err := c.client.Scan()
	if err != nil {
		log.Fatal(err)
	}

	ipv6Addr, err := c.client.LL64(pd.Addr)
	if err != nil {
		log.Fatal(err)
	}

	pd.IPV6Addr = ipv6Addr
	log.Printf("Translated address:%#v", pd)

	err = c.client.SRegS2(pd.Channel)
	if err != nil {
		log.Fatal(err)
	}

	err = c.client.SRegS3(pd.PanID)
	if err != nil {
		log.Fatal(err)
	}

	joined, err := c.client.Join(pd)
	if err != nil {
		log.Fatal(err)
	}

	if !joined {
		log.Fatal("Failed to join")
	}

	// TODO: return error
	return nil
}

// GetCurrentPowerConsumption is ..
func (c Controller) GetCurrentPowerConsumption(ctx context.Context) (int ,error) {
	elframe := []byte{0x10, 0x81, 0x00, 0x01, 0x05, 0xff, 0x01, 0x02, 0x88, 0x01, 0x62, 0x01, 0xe7, 0x00}
	ipv6addr := "FE80:0000:0000:0000:021C:6400:030C:12A4"

	eldata, err := c.client.SendTo(ipv6addr, elframe)
	if err != nil {
		log.Println(err)
	}
	elFrame, err := echonetlite.ParseFrame(eldata)
	if err != nil {
		return 0, fmt.Errorf("invalid frame: %w", err)
	}
	elFrame.Print()

	return 0, nil
}


