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

// Object is object
type Object struct {
	Data Data
}

// Class returns Class
func (o Object) Class() Class {
	return Class{o.Data[0], o.Data[1]}
}

func (o Object) classGroupCode() ClassGroupCode {
	return ClassGroupCode(o.Data[0])
}

func (o Object) classCode() ClassCode {
	return ClassCode(o.Data[1])
}

func (o Object) isNodeProfile() bool {
	if o.Data[0] == byte(ProfileGroup) &&
		o.Data[1] == byte(Profile) {
		return true
	}
	return false
}

// Frame is Echonet-Lite frame
type Frame struct {
	Data       Data    // Entire Data
	EHD        Data    // Echonet Lite Header
	TID        Data    // Transaction ID
	EData      Data    // Echonet Lite Data
	SEOJ       Object  // Source Echonet Lite Object
	DEOJ       Object  // Destination Echonet Lite Object
	ESV        ESVType // Echonet Lite Service
	OPC        Data    // Num of Properties
	Properties []Property
	Object     interface{}
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
	SEOJ := Object{Data: EDATA[:3]}
	DEOJ := Object{Data: EDATA[3:6]}
	ESV := ESVType(EDATA[6:7][0])
	OPC := EDATA[7:8]

	pNum := int(OPC[0])
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

	f := Frame{Data: frame, EHD: EHD, TID: TID, EData: EDATA, SEOJ: SEOJ, DEOJ: DEOJ, ESV: ESV, OPC: OPC, Properties: props}
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

	str := fmt.Sprintf("%s EHD[%s] TID[%s] SEOJ[%s](%s) DEOJ[%s](%s) ESV[%s] OPC[%s]", f.Data, f.EHD, f.TID, f.SEOJ, sObjInfo.Desc, f.DEOJ, dObjInfo.Desc, f.ESV, f.OPC)
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
		logger.Println(" ", f.Data)
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
func CreateInfFrame(transid uint16) *Frame {
	// INF
	data := []byte{0x10, 0x81, byte(transid >> 8 & 0xFF), byte(transid & 0xFF), 0x0e, 0xf0, 0x01, 0x0e, 0xf0, 0x01, 0x73, 0x01, 0xd5, 0x04, 0x01, 0x05, 0xff, 0x01}
	//data := []byte{0x10, 0x81, 0x0, 0x0, 0x05, 0xff, 0x01, 0x0e, 0xf0, 0x01, 0x63, 0x01, 0xd5, 0x00}
	frame, err := ParseFrame(data)
	if err != nil {
		logger.Print("Error:", err)
		return nil
	}
	//logger.Print(frame)
	return &frame
}

// CreateInfReqFrame creates INF_REQ frame
func CreateInfReqFrame(transid uint16) *Frame {
	// INF_REQ
	data := []byte{0x10, 0x81, byte(transid >> 8 & 0xFF), byte(transid & 0xFF), 0x05, 0xff, 0x01, 0x0e, 0xf0, 0x01, 0x63, 0x01, 0xd5, 0x00}
	frame, err := ParseFrame(data)
	if err != nil {
		logger.Print("Error:", err)
		return nil
	}
	//logger.Print(frame)
	return &frame
}

// CreateGetFrame creates GET frame
func CreateGetFrame(transid uint16) *Frame {
	// Get
	//data := []byte{0x10, 0x81, 0x0, 0x0, 0x05, 0xff, 0x01, 0x0e, 0xf0, 0x01, 0x62, 0x01, 0xd6, 0x00}
	data := []byte{0x10, 0x81, byte(transid >> 8 & 0xFF), byte(transid & 0xFF), 0x05, 0xff, 0x01, 0x0e, 0xf0, 0x01, 0x62, 0x08, 0x80, 0x00, 0x82, 0x00, 0xd3, 0x00, 0xd4, 0x00, 0xd5, 0x00, 0xd6, 0x00, 0xd7, 0x00, 0x9f, 0x00}
	frame, err := ParseFrame(data)
	if err != nil {
		logger.Print("Error:", err)
		return nil
	}
	//logger.Print(frame)
	return &frame
}

// CreateAirconGetFrame creates GET air-con info frame
// TODO: refactor
func CreateAirconGetFrame(transid uint16) *Frame {
	// Get
	//data := []byte{0x10, 0x81, 0x0, 0x0, 0x05, 0xff, 0x01, 0x01, 0x30, 0x01, 0x62, 0x04, 0x81, 0x00, 0x83, 0x00, 0xbb, 0x00, 0xbe, 0x00}
	data := make([]byte, 0)

	ehd1 := []byte{byte(EchonetLite)} // 10
	ehd2 := []byte{byte(FixedFormat)} // 81
	tid := []byte{byte(transid >> 8 & 0xFF), byte(transid & 0xFF)}

	edata := make([]byte, 0)
	seoj := []byte{byte(ControllerGroup), byte(Controller), 0x01}             // 05FF01
	deoj := []byte{byte(AirConditionerGroup), byte(HomeAirConditioner), 0x01} // 013001
	esv := []byte{byte(Get)}                                                  // 62

	properties := []struct {
		epc []byte
		edt []byte
	}{
		{epc: []byte{byte(InstallationLocation)}, edt: []byte{}},       //0x81
		{epc: []byte{byte(ID)}, edt: []byte{}},                         //0x83
		{epc: []byte{byte(MeasuredRoomTemperature)}, edt: []byte{}},    // 0xBB
		{epc: []byte{byte(MeasuredOutdoorTemperature)}, edt: []byte{}}, //0xBE
	}

	edata = append(edata, seoj...)
	edata = append(edata, deoj...)
	edata = append(edata, esv...)

	opc := byte(len(properties))
	edata = append(edata, opc)

	for _, p := range properties {
		edata = append(edata, p.epc...)
		edata = append(edata, byte(len(p.edt)))
		edata = append(edata, p.edt...)
	}

	data = append(data, ehd1...)
	data = append(data, ehd2...)
	data = append(data, tid...)
	data = append(data, edata...)

	frame, err := ParseFrame(data)
	if err != nil {
		logger.Print("Error:", err)
		return nil
	}
	//logger.Print(frame)
	return &frame
}
