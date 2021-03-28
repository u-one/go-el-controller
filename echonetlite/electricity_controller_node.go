package echonetlite

import (
	"encoding/binary"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	gpower = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "home",
			Subsystem: "smartmeter_exporter",
			Name:      "instantpower",
			Help:      "",
		},
	)
)

func init() {
	prometheus.MustRegister(gpower)
}

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

	rdata, err := n.client.Send(f.Serialize())
	if err != nil {
		return 0, err
	}
	rf, err := ParseFrame(rdata)
	if err != nil {
		return 0, fmt.Errorf("invalid frame: %w", err)
	}
	rf.Print()

	switch rf.ESV {
	// 応答・通知
	case GetRes: // プロパティ値読み出し応答
		o := rf.SrcObj()
		switch o.classGroupCode() {
		case HomeEquipmentGroup:
			switch o.classCode() {
			case LowVoltageSmartMeter:
				for _, p := range rf.Properties {
					switch PropertyCode(p.Code) {
					case InstantPower:
						power := binary.BigEndian.Uint32(p.Data)
						gpower.Set(float64(power))
						logger.Printf("Power: %d [W]", power)
						return int(power), nil
					}
				}
			}
		}
	default:
	}

	return 0, nil
}

// CreateCurrentPowerConsumptionFrame creates GET current power consumption frame
func CreateCurrentPowerConsumptionFrame(transID uint16) *Frame {
	// Get
	src := NewObject(ControllerGroup, Controller, 0x01)
	dest := NewObject(HomeEquipmentGroup, LowVoltageSmartMeter, 0x01)

	props := []Property{}
	props = append(props, Property{Code: byte(InstantPower), Len: 0, Data: []byte{}})

	frame := NewFrame(transID, src, dest, Get, props)
	return &frame
}
