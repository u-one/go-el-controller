package hems

import (
	"fmt"

	"github.com/u-one/go-el-controller/wisun"
)

// ElectricityControllerNode is node for smart-meter
type ElectricityControllerNode struct {
	client wisun.Client
}

// NewElectricityControllerNode returns ElectricityControllerNode instance
func NewElectricityControllerNode(c wisun.Client) *ElectricityControllerNode {
	return &ElectricityControllerNode{c}
}

// Close closes client
func (n ElectricityControllerNode) Close() {
	n.client.Close()
}

// Start starts to connect to smart-meter
func (n ElectricityControllerNode) Start(bRouteID, bRoutePassword string) error {
	err := n.client.Connect(bRouteID, bRoutePassword)
	if err != nil {
		return fmt.Errorf("PANA authentication failed: %v", err)
	}
	return nil
}

// GetPowerConsumption requests power consumption and receives
func (n ElectricityControllerNode) GetPowerConsumption() (int, error) {
	return n.client.Get()
}
