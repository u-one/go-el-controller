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
	Object     interface{}
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
	eData = append(eData, f.SEOJ.Data...)
	eData = append(eData, f.DEOJ.Data...)
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
		EDT := EDATA[epcOffset+2 : epcOffset+2+propertyValueLen]

		props = append(props, Property{Code: EPC[0], Len: int(PDC[0]), Data: EDT})

		epcOffset += (2 + propertyValueLen)
	}

	f := Frame{EHD: EHD, TID: TID, SEOJ: SEOJ, DEOJ: DEOJ, ESV: ESV, OPC: OPC, Properties: props}
	err := f.parseFrame()
	return f, err
}

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

func (f *Frame) parseFrame() error {
	class := f.SrcClass()
	//logger.Println("frameReceived:", class)

	switch ClassGroupCode(class.ClassGroupCode) {
	case AirConditionerGroup:
		switch ClassCode(class.ClassCode) {
		case HomeAirConditioner:
			logger.Println("エアコン")
			obj := AirconObject{}
			for _, p := range f.Properties {
				logger.Printf("Property Code: %x, %#v\n", p.Code, p.Data)
				switch PropertyCode(p.Code) {
				case OperationStatus:
				case InstallationLocation:
					if p.Len != 1 {
						return fmt.Errorf("InstallationLocation invalid length: %d", p.Len)
					}
					var d byte = p.Data[0]
					logger.Printf("%08b\n", d)
					if d>>7 == 1 {
						// free definition
						logger.Println("free definition")
						break
					}
					locationCode := (d >> 3) & 0x0F
					locationNo := d & 0x07
					obj.InstallLocation = Location{Code: LocationCode(locationCode), Number: int32(locationNo)}
					logger.Printf("locationCode: %0b locationNo: %0b\n", locationCode, locationNo)
				case ID:
					if p.Len == 0 {
						return fmt.Errorf("ID invalid length: %d", p.Len)
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
				case MeasuredRoomTemperature:
					if p.Len != 1 {
						return fmt.Errorf("MeasuredRoomTemperature invalid length: %d", p.Len)
					}
					temp := int(p.Data[0])
					obj.InternalTemp = float64(temp)
					logger.Printf("室温:%d℃\n", temp)
					break
				case MeasuredOutdoorTemperature:
					if p.Len != 1 {
						return fmt.Errorf("MeasuredOutdoorTemperature invalid length: %d", p.Len)
					}
					temp := int(p.Data[0])
					obj.OuterTemp = float64(temp)
					logger.Printf("外気温:%d℃\n", temp)
					break
				}
			}
			f.Object = obj
			break
		}
		break
	}
	return nil

}

// SrcClass returns src class
func (f Frame) SrcClass() Class {
	return NewClass(f.SEOJ)
}

// DstClass returns dst class
func (f Frame) DstClass() Class {
	return NewClass(f.DEOJ)
}

// String returns string
func (f Frame) String() string {
	sObjInfo := ClassInfoDB.Get(f.SrcClass())
	dObjInfo := ClassInfoDB.Get(f.DstClass())

	str := fmt.Sprintf("%s EHD[%s] TID[%s] SEOJ[%s](%s) DEOJ[%s](%s) ESV[%s] OPC[%d]", f.Serialize(), f.EHD, f.TID, f.SEOJ, sObjInfo.Desc, f.DEOJ, dObjInfo.Desc, f.ESV, f.OPC)
	for i, p := range f.Properties {
		desc := ""
		if sObjInfo != nil {
			prop := sObjInfo.Properties[PropertyCode(p.Code)]
			if prop != nil {
				desc = prop.Detail
			}
		}

		str = str + fmt.Sprintf(" EPC%d[%x](%s)", i, p.Code, desc)
		str = str + fmt.Sprintf(" PDC%d[%d]", i, p.Len)
		str = str + fmt.Sprintf(" EDT%d[%s]", i, p.Data)
	}
	return str
}

// Print prints frame detail
func (f Frame) Print() {
	sObjInfo := ClassInfoDB.Get(f.SrcClass())
	dObjInfo := ClassInfoDB.Get(f.DstClass())

	if false {
		logger.Println("============ Frame ============")
		logger.Println(" ", f.Serialize())
		logger.Println("  -----------------------------")
		logger.Println("  EHD:", f.EHD)
		logger.Println("  TID:", f.TID)
		logger.Println("  SEOJ:", f.SEOJ, " (Source Echonet lite object)", sObjInfo.Desc)
		logger.Println("  DEOJ:", f.DEOJ, " (Dest. Echonet lite object)", dObjInfo.Desc)
		logger.Println("  ESV:", f.ESV, " (Echonet Lite Service)")
		logger.Println("  OPC:", f.OPC, " (Num of properties)")

		for i, p := range f.Properties {
			desc := ""
			if sObjInfo != nil {
				prop := sObjInfo.Properties[PropertyCode(p.Code)]
				if prop != nil {
					desc = prop.Detail
				}
			}

			logger.Printf("  EPC%d: %x (Echonet lite property) %s", i, p.Code, desc)
			logger.Printf("  PDC%d: %d (Length (bytes) of EDT)", i, p.Len)
			logger.Printf("  EDT%d: %s (Property data)", i, p.Data)
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
