package hems

import (
	"fmt"

	"github.com/u-one/go-el-controller/wisun"
)

// ElectricityMeterClient is client for electricity meter
type ElectricityMeterClient struct {
	controller *wisun.Controller
}

// NewElectricityMeterClient returns ElectricityMeterClient instance
func NewElectricityMeterClient(c wisun.Client) *ElectricityMeterClient {
	return &ElectricityMeterClient{
		wisun.NewController(c),
	}
}

// Close closes client
func (c ElectricityMeterClient) Close() {
	c.controller.Close()
}

// Start starts sequence
func (c ElectricityMeterClient) Start(bRouteID, bRoutePassword string) error {
	err := c.controller.PANAAuth(bRouteID, bRoutePassword)
	if err != nil {
		return fmt.Errorf("PANA authentication failed: %v", err)
	}
	defer c.controller.Close()
	return nil
}

// Version requests
func (c ElectricityMeterClient) Version() {
	c.controller.Version()

}
