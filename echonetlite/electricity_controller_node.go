package echonetlite

import (
	"fmt"
)

// SmartMeterClient is interface for smart-meter cleint
type SmartMeterClient interface {
	Connect(bRouteID, bRoutePW string) error
	Close()
	Send(data []byte) ([]byte, error)
}

// ElectricityControllerNode is node for smart-meter
type ElectricityControllerNode struct {
	client SmartMeterClient
}

// NewElectricityControllerNode returns ElectricityControllerNode instance
func NewElectricityControllerNode(c SmartMeterClient) *ElectricityControllerNode {
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
		return fmt.Errorf("Connect failed: %v", err)
	}
	return nil
}

// GetPowerConsumption requests power consumption and receives
func (n ElectricityControllerNode) GetPowerConsumption() (int, error) {
	f := CreateCurrentPowerConsumptionFrame(1) // TODO: increment

	eldata, err := n.client.Send(f.Serialize())
	if err != nil {
		return 0, err
	}
	elFrame, err := ParseFrame(eldata)
	if err != nil {
		return 0, fmt.Errorf("invalid frame: %w", err)
	}
	elFrame.Print()
	return 0, nil
}
