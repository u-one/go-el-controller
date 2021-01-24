package echonetlite

// ClassGroupCode represents class gruop code
type ClassGroupCode byte

// ClassCode represents class code
type ClassCode byte

// definition of class group codes
const (
	SensorGroup         ClassGroupCode = 0x00
	AirConditionerGroup ClassGroupCode = 0x01
	HomeEquipmentGroup  ClassGroupCode = 0x02
	HomeApplianceGroup  ClassGroupCode = 0x03
	HealthCareGroup     ClassGroupCode = 0x04
	ControllerGroup     ClassGroupCode = 0x05
	AVGroup             ClassGroupCode = 0x06

	ProfileGroup ClassGroupCode = 0x0E
)

// Profile is definition of profile object class code
const Profile ClassCode = 0xF0

// definition of class codes for AirConditionerGroup
const (
	HomeAirConditioner ClassCode = 0x30
)

// definition of class codes for HomeEquipmentGroup
const (
	LowVoltageSmartMeter ClassCode = 0x88
)

// definition of class codes for ControllerGroup
const (
	HandHeldDevice ClassCode = 0xFE
	Controller     ClassCode = 0xFF
)
