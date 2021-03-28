package echonetlite

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "[Frame]", log.LstdFlags)
	//logger = log.New(ioutil.Discard, "[Frame]", log.LstdFlags)

}

// Data represents binary data
type Data []byte

// String returns hex representation for the data
func (d Data) String() string {
	return hex.EncodeToString(d)
}

// EHD1
const (
	EchonetLite byte = 0x10
)

// EHD2
const (
	FixedFormat     byte = 0x81
	ArbitraryFormat byte = 0x82
)

// Frame is Echonet-Lite frame
type Frame struct {
	EHD        Data    // Echonet Lite Header
	TID        Data    // Transaction ID
	SEOJ       Object  // Source Echonet Lite Object
	DEOJ       Object  // Destination Echonet Lite Object
	ESV        ESVType // Echonet Lite Service
	OPC        byte    // Num of Properties
	Properties []Property
}

// NewFrame retunrs Frame
func NewFrame(transID uint16, src, dest Object, service ESVType, props []Property) Frame {

	ehd := []byte{byte(EchonetLite), byte(FixedFormat)}
	tid := []byte{byte(transID >> 8 & 0xFF), byte(transID & 0xFF)}
	seoj := src
	deoj := dest
	esv := service

	f := Frame{EHD: ehd, TID: tid, SEOJ: seoj, DEOJ: deoj, ESV: esv, OPC: byte(len(props)), Properties: props}

	return f
}

// EData returns serialized EDATA part
func (f Frame) EData() Data {
	eData := []byte{}
	eData = append(eData, f.SEOJ.Data()...)
	eData = append(eData, f.DEOJ.Data()...)
	eData = append(eData, byte(f.ESV))
	eData = append(eData, byte(f.OPC))
	for _, p := range f.Properties {
		eData = append(eData, p.Serialize()...)
	}
	return eData
}

// Serialize returns its serialized data
func (f Frame) Serialize() Data {
	frame := []byte{}
	frame = append(frame, f.EHD...)
	frame = append(frame, f.TID...)
	frame = append(frame, f.EData()...)
	return frame
}

// ParseFrame returns Frame
func ParseFrame(data []byte) (Frame, error) {
	if len(data) < 9 {
		return Frame{}, fmt.Errorf("size is too short:%d", len(data))
	}
	frame := Data(data)
	EHD := frame[:2]
	TID := frame[2:4]
	EDATA := frame[4:]
	SEOJ := NewObjectFromData(EDATA[:3])
	DEOJ := NewObjectFromData(EDATA[3:6])
	ESV := ESVType(EDATA[6:7][0])
	OPC := EDATA[7:8][0]

	pNum := int(OPC)
	props := make([]Property, 0, pNum)

	epcOffsetBase := 8
	epcOffset := epcOffsetBase
	for i := 0; i < pNum; i++ {
		EPC := EDATA[epcOffset : epcOffset+1]
		PDC := EDATA[epcOffset+1 : epcOffset+2]
		propertyValueLen := int(PDC[0])
		if len(EDATA) < epcOffset+2+propertyValueLen {
			return Frame{}, fmt.Errorf("invalid EDT length")
		}
		EDT := EDATA[epcOffset+2 : epcOffset+2+propertyValueLen]

		props = append(props, Property{Code: EPC[0], Len: int(PDC[0]), Data: EDT})

		epcOffset += (2 + propertyValueLen)
	}

	f := Frame{EHD: EHD, TID: TID, SEOJ: SEOJ, DEOJ: DEOJ, ESV: ESV, OPC: OPC, Properties: props}
	return f, nil
}

// parseProperties parses properties
func parseProperties(obj Object, properties []Property) (interface{}, error) {
	logger.Printf("ParseProperties: %v", properties)

	switch obj.classGroupCode() {
	case ProfileGroup:
		switch obj.classCode() {
		case Profile:
			logger.Println("ノードプロファイル")
			for _, p := range properties {
				parseNodeProfileProperty(p)
			}
			return nil, nil
		}
	case AirConditionerGroup:
		switch obj.classCode() {
		case HomeAirConditioner:
			logger.Println("エアコン")
			obj := AirconObject{}
			for _, p := range properties {
				parseHomeAirConditionerProperty(p, &obj)
			}
			return obj, nil
		}
		break
	}
	return nil, nil
}

// TODO: Unify Object struct and structs below

// SuperObject is super object
type SuperObject struct {
	InstallLocation Location
}

// AirconObject is object for aircon
type AirconObject struct {
	SuperObject
	InternalTemp float64
	OuterTemp    float64
}

func parseSuperObjectProperty(p Property) bool {
	logger.Printf("Super Class Property Code: %x, %#v\n", p.Code, p.Data)
	if len(p.Data) == 0 {
		return false
	}

	switch PropertyCode(p.Code) {
	case OperationStatus: // 0x80
		data := p.Data[0]
		if data == 0x30 {
			logger.Println("ON")
		} else if data == 0x31 {
			logger.Println("OFF")
		} else {
			logger.Println("UNKNOWN")
		}
		return true
	case SpecVersion: // 0x82
		logger.Printf("SpecVersion: %c", rune(p.Data[2]))
		return true
	case ID: // 0x83
		logger.Printf("ID: %x", p.Data)
		return true
	case NumOfInstances:
		return true
	case NumOfClasses:
		return true
	case InstanceListNotification:
		return true
	case InstanceListS:
		return true
	case ClassListS:
		return true
	case GetPropertyMap: //0x9F
		return true
	}
	return false
}

func parseNodeProfileProperty(p Property) bool {
	if parseSuperObjectProperty(p) {
		return true
	}

	logger.Printf("Property Code: %x, %#v\n", p.Code, p.Data)
	switch PropertyCode(p.Code) {
	case NumOfInstances: // 0xD3
		logger.Printf("Num of instances: %x", p.Data)
		return true
	case NumOfClasses: // 0xD4
		logger.Printf("Num of classes: %x", p.Data)
		return true
	case InstanceListNotification: // 0xD5
		var instances int
		if len(p.Data) > 0 {
			instances = int(p.Data[0])
		}
		var objCode Data
		if len(p.Data) > 1 {
			objCode = p.Data[1:]
		}
		logger.Printf("Num of notification instances: %x, Object code: %x", instances, objCode)
		return true
	case InstanceListS: // 0xD6
		logger.Printf("Num of instances S: %x, Object code: %v", p.Data[0], p.Data[1:len(p.Data)])
		return true
	case ClassListS: // 0xD7
		logger.Printf("Num of classes S: %x, Object code: %v", p.Data[0], p.Data[1:len(p.Data)])
		return true
	}
	return false
}

func parseHomeAirConditionerProperty(p Property, obj *AirconObject) bool {
	if parseSuperObjectProperty(p) {
		return true
	}

	logger.Printf("Property Code: %x, %#v\n", p.Code, p.Data)
	switch PropertyCode(p.Code) {
	case OperationStatus:
		return true
	case InstallationLocation:
		if p.Len != 1 {
			logger.Printf("[Error] InstallationLocation invalid length: %d", p.Len)
			return true
		}
		var d byte = p.Data[0]
		logger.Printf("%08b\n", d)
		if d>>7 == 1 {
			// free definition
			logger.Println("free definition")
			return true
		}
		locationCode := (d >> 3) & 0x0F
		locationNo := d & 0x07
		obj.InstallLocation = Location{Code: LocationCode(locationCode), Number: int32(locationNo)}
		logger.Printf("locationCode: %0b locationNo: %0b\n", locationCode, locationNo)
		return true
	case ID:
		if p.Len == 0 {
			logger.Printf("[Error] ID invalid length: %d", p.Len)
			return true
		}
		lowerCommunicationLayerID := p.Data[0]
		switch {
		case 0x00 == lowerCommunicationLayerID:
		case 0xFE > lowerCommunicationLayerID:
		case 0xFE == lowerCommunicationLayerID:
			manufacturerCode := p.Data[1:4]
			manufacturerID := p.Data[4:]
			logger.Printf("メーカコード: %#v メーカID: %#v\n", manufacturerCode, manufacturerID)
		case 0xFF == lowerCommunicationLayerID:
		}
		return true
	case MeasuredRoomTemperature:
		if p.Len != 1 {
			logger.Printf("[Error] MeasuredRoomTemperature invalid length: %d", p.Len)
			return true
		}
		temp := int(p.Data[0])
		obj.InternalTemp = float64(temp)
		logger.Printf("室温:%d℃\n", temp)
		return true
	case MeasuredOutdoorTemperature:
		if p.Len != 1 {
			logger.Printf("[Error] MeasuredOutdoorTemperature invalid length: %d", p.Len)
			return true
		}
		temp := int(p.Data[0])
		obj.OuterTemp = float64(temp)
		logger.Printf("外気温:%d℃\n", temp)
		return true
	}
	return false
}

// SrcObj returns src Object
func (f Frame) SrcObj() Object {
	return f.SEOJ
}

// DstObj returns dst Object
func (f Frame) DstObj() Object {
	return f.DEOJ
}

// String returns string
func (f Frame) String() string {
	str := fmt.Sprintf("%s EHD[%s] TID[%s] SEOJ[%s] DEOJ[%s] ESV[%s] OPC[%d]", f.Serialize(), f.EHD, f.TID, f.SEOJ, f.DEOJ, f.ESV, f.OPC)
	for i, p := range f.Properties {
		str = str + fmt.Sprintf(" %d %s", i, p)
	}
	return str
}

// Print prints frame detail
func (f Frame) Print() {
	if false {
		logger.Println("============ Frame ============")
		logger.Println(" ", f.Serialize())
		logger.Println("  -----------------------------")
		logger.Println("  EHD:", f.EHD)
		logger.Println("  TID:", f.TID)
		logger.Println("  SEOJ:", f.SEOJ, " (Source Echonet lite object)")
		logger.Println("  DEOJ:", f.DEOJ, " (Dest. Echonet lite object)")
		logger.Println("  ESV:", f.ESV, " (Echonet Lite Service)")
		logger.Println("  OPC:", f.OPC, " (Num of properties)")

		for i, p := range f.Properties {
			logger.Printf(" %d %s", i, p)
		}
		logger.Println("===============================")
	} else {
		logger.Println(f)
	}
}

/*
func (f Frame) String() string {
	return hex.EncodeToString(f.Data)
}
*/

// EPCCode is EPC code
type EPCCode byte

// CreateInfFrame creates INF frame
func CreateInfFrame(transID uint16) *Frame {
	// INF
	src := NewObject(ProfileGroup, Profile, 0x01)
	dest := NewObject(ProfileGroup, Profile, 0x01)

	props := []Property{}
	props = append(props, Property{Code: 0xd5, Len: 4, Data: []byte{0x01, 0x05, 0xff, 0x01}})
	frame := NewFrame(transID, src, dest, Inf, props)
	return &frame
}

// CreateInfReqFrame creates INF_REQ frame
func CreateInfReqFrame(transID uint16) *Frame {
	// INF_REQ

	src := NewObject(ControllerGroup, Controller, 0x01)
	dest := NewObject(ProfileGroup, Profile, 0x01)

	props := []Property{}
	props = append(props, Property{Code: 0xd5, Len: 0, Data: []byte{}})
	frame := NewFrame(transID, src, dest, InfReq, props)
	return &frame
}

// CreateGetFrame creates GET frame
func CreateGetFrame(transID uint16) *Frame {
	// Get
	src := NewObject(ControllerGroup, Controller, 0x01)
	dest := NewObject(ProfileGroup, Profile, 0x01)

	props := []Property{}
	props = append(props, Property{Code: byte(OperationStatus), Len: 0, Data: []byte{}})
	props = append(props, Property{Code: byte(SpecVersion), Len: 0, Data: []byte{}})
	props = append(props, Property{Code: byte(NumOfInstances), Len: 0, Data: []byte{}})
	props = append(props, Property{Code: byte(NumOfClasses), Len: 0, Data: []byte{}})
	props = append(props, Property{Code: byte(InstanceListNotification), Len: 0, Data: []byte{}})
	props = append(props, Property{Code: byte(InstanceListS), Len: 0, Data: []byte{}})
	props = append(props, Property{Code: byte(ClassListS), Len: 0, Data: []byte{}})
	props = append(props, Property{Code: 0x9f, Len: 0, Data: []byte{}})

	frame := NewFrame(transID, src, dest, Get, props)
	return &frame
}

// CreateAirconGetFrame creates GET air-con info frame
func CreateAirconGetFrame(transID uint16) *Frame {
	// Get
	src := NewObject(ControllerGroup, Controller, 0x01)
	dest := NewObject(AirConditionerGroup, HomeAirConditioner, 0x01)

	props := []Property{}
	props = append(props, Property{Code: byte(InstallationLocation), Len: 0, Data: []byte{}})
	props = append(props, Property{Code: byte(ID), Len: 0, Data: []byte{}})
	props = append(props, Property{Code: byte(MeasuredRoomTemperature), Len: 0, Data: []byte{}})
	props = append(props, Property{Code: byte(MeasuredOutdoorTemperature), Len: 0, Data: []byte{}})

	frame := NewFrame(transID, src, dest, Get, props)
	return &frame
}
