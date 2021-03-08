package wisun

import (
	"fmt"
	"log"

	"github.com/u-one/go-el-controller/echonetlite"
)

// Controller is
type Controller struct {
	client  Client
	panDesc PanDesc
}

// NewController returns new instance
func NewController(c Client) *Controller {
	return &Controller{client: c}
}

// Close close
func (c Controller) Close() {
	c.client.Close()
}

// Version returns software version that installed in the client
func (c Controller) Version() error {
	return c.client.Version()
}

// PANAAuth starts PANA authentication
func (c *Controller) PANAAuth(bRouteID, bRoutePW string) error {

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

	c.panDesc = pd

	// TODO: return error
	return nil
}

// GetCurrentPowerConsumption is ..
func (c Controller) GetCurrentPowerConsumption() (int, error) {
	f := echonetlite.CreateCurrentPowerConsumptionFrame(1) // TODO: increment

	eldata, err := c.client.SendTo(c.panDesc.IPV6Addr, f.Serialize())
	if err != nil {
		return 0, err
	}
	elFrame, err := echonetlite.ParseFrame(eldata)
	if err != nil {
		return 0, fmt.Errorf("invalid frame: %w", err)
	}
	elFrame.Print()

	return 0, nil
}
