package hems

import (
	"fmt"

	"github.com/u-one/go-el-controller/wisun"
)

// ElectricityMeterClient is client for electricity meter
type ElectricityMeterClient struct {
	client wisun.Client
}

// NewElectricityMeterClient returns ElectricityMeterClient instance
func NewElectricityMeterClient(c wisun.Client) *ElectricityMeterClient {
	return &ElectricityMeterClient{c}
}

// Close closes client
func (c ElectricityMeterClient) Close() {
	c.client.Close()
}

// Start starts sequence
func (c ElectricityMeterClient) Start(bRouteID, bRoutePassword string) error {
	err := c.client.Connect(bRouteID, bRoutePassword)
	if err != nil {
		return fmt.Errorf("PANA authentication failed: %v", err)
	}
	return nil
}

// GetPowerConsumption requests power consumption and receives
func (c ElectricityMeterClient) GetPowerConsumption() (int, error) {
	return c.client.Get()
}
