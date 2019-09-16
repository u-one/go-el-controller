package main

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/u-one/go-el-controller/class"
)

// Data represents binary data
type Data []byte

// String returns hex representation for the data
func (d Data) String() string {
	return hex.EncodeToString(d)
}

// Property represents Echonet-Lite property
type Property struct {
	Code byte
	Len  int
	Data Data
}

// Frame is Echonet-Lite frame
type Frame struct {
	Data       Data    // Entire Data
	EHD        Data    // Echonet Lite Header
	TID        Data    // Transaction ID
	EData      Data    // Echonet Lite Data
	SEOJ       Data    // Source Echonet Lite Object
	DEOJ       Data    // Destination Echonet Lite Object
	ESV        ESVType // Echonet Lite Service
	OPC        Data    // Num of Properties
	Properties []Property
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
	SEOJ := EDATA[:3]
	DEOJ := EDATA[3:6]
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

	return Frame{Data: frame, EHD: EHD, TID: TID, EData: EDATA, SEOJ: SEOJ, DEOJ: DEOJ, ESV: ESV, OPC: OPC, Properties: props}, nil
}

// SrcClassCode returns src class code
func (f Frame) SrcClassCode() class.ClassCode {
	return class.NewClassCode(f.SEOJ[0], f.SEOJ[1])
}

// DstClassCode returns dst class code
func (f Frame) DstClassCode() class.ClassCode {
	return class.NewClassCode(f.DEOJ[0], f.DEOJ[1])
}

// Print prints frame detail
func (f Frame) Print() {
	sObjInfo := ClassInfoMap.Get(f.SrcClassCode())
	dObjInfo := ClassInfoMap.Get(f.DstClassCode())

	log.Println("============ Frame ============")
	log.Println(" ", f.Data)
	log.Println("  -----------------------------")
	log.Println("  EHD:", f.EHD)
	log.Println("  TID:", f.TID)
	log.Println("  SEOJ:", f.SEOJ, " (Source Echonet lite object)", sObjInfo.Desc)
	log.Println("  DEOJ:", f.DEOJ, " (Dest. Echonet lite object)", dObjInfo.Desc)
	log.Println("  ESV:", f.ESV, " (Echonet Lite Service)")
	log.Println("  OPC:", f.OPC, " (Num of properties)")

	for i, p := range f.Properties {
		desc := ""
		if sObjInfo != nil {
			prop := sObjInfo.Properties[class.PropertyCode(p.Code)]
			if prop != nil {
				desc = prop.Detail
			}
		}

		log.Printf("  EPC%d: %x (Echonet lite property) %s", i, p.Code, desc)
		log.Printf("  PDC%d: %d (Length (bytes) of EDT)", i, p.Len)
		log.Printf("  EDT%d: %s (Property data)", i, p.Data)
	}
	log.Println("===============================")
}

// ESVType represnts type of ESV
type ESVType byte

// ESVTypes
const (
	SetI   ESVType = 0x60 // SetI
	SetC   ESVType = 0x61 // SetC
	Get    ESVType = 0x62 // Get
	InfReq ESVType = 0x63 // INF_REQ
	SetGet ESVType = 0x6E // SetGet

	SetRes    ESVType = 0x71 // Set_Res
	GetRes    ESVType = 0x72 // Get_Res
	Inf       ESVType = 0x73 // INF
	InfC      ESVType = 0x74 // INFC
	InfCRes   ESVType = 0x7A // INFC_Res
	SetGetRes ESVType = 0x7E // SetGet_Res

	SetISNA   ESVType = 0x50 // SetI_SNA
	SetCSNA   ESVType = 0x51 // SetC_SNA
	GetSNA    ESVType = 0x52 // Get_SNA
	InfSNA    ESVType = 0x53 // INF_SNA
	SetGetSNA ESVType = 0x5E // SetGet_SNA
)

func (t ESVType) String() string {
	switch t {
	case SetI:
		return "SetI"
	case SetC:
		return "SetC"
	case Get:
		return "Get"
	case InfReq:
		return "INF_REQ"
	case SetGet:
		return "SetGet"
	case SetRes:
		return "Set_Res"
	case GetRes:
		return "Get_Res"
	case Inf:
		return "INF"
	case InfC:
		return "INFC"
	case InfCRes:
		return "INFC_Res"
	case SetGetRes:
		return "SetGet_Res"
	case SetISNA:
		return "SetI_SNA"
	case SetCSNA:
		return "SetC_SNA"
	case GetSNA:
		return "Get_SNA"
	case InfSNA:
		return "INF_SNA"
	case SetGetSNA:
		return "SetGet_SNA"
	default:
		return "UNKNOWN" + hex.EncodeToString([]byte{byte(t)})
	}
}

func (f Frame) String() string {
	return hex.EncodeToString(f.Data)
}

// ClassGroupCode represents class gruop code
type ClassGroupCode byte

// ClassCode represents class code
type ClassCode byte

// PropertyCode represents property code
type PropertyCode byte

// EPCCode is EPC code
type EPCCode byte

// PropertyDefs is a map EPCCode as key detail string as value
type PropertyDefs map[EPCCode]string

// definition of class group codes
const (
	AirConditioner ClassGroupCode = 0x01
)

// definition of class codes
const (
	HomeAirConditioner ClassCode = 0x30
)

// definition of property codes
const (
	MeasuredRoomTemperature    PropertyCode = 0xBB
	MeasuredOutdoorTemperature PropertyCode = 0xBE
)

func createInfFrame() *Frame {
	// INF
	data := []byte{0x10, 0x81, 0x0, 0x0, 0x0e, 0xf0, 0x01, 0x0e, 0xf0, 0x01, 0x73, 0x01, 0xd5, 0x04, 0x01, 0x05, 0xff, 0x01}
	//data := []byte{0x10, 0x81, 0x0, 0x0, 0x05, 0xff, 0x01, 0x0e, 0xf0, 0x01, 0x63, 0x01, 0xd5, 0x00}
	frame, err := ParseFrame(data)
	if err != nil {
		log.Print("Error:", err)
		return nil
	}
	log.Print(frame)
	return &frame
}

func createInfReqFrame() *Frame {
	// INF_REQ
	data := []byte{0x10, 0x81, 0x0, 0x0, 0x05, 0xff, 0x01, 0x0e, 0xf0, 0x01, 0x63, 0x01, 0xd5, 0x00}
	frame, err := ParseFrame(data)
	if err != nil {
		log.Print("Error:", err)
		return nil
	}
	log.Print(frame)
	return &frame
}

func createGetFrame() *Frame {
	// Get
	//data := []byte{0x10, 0x81, 0x0, 0x0, 0x05, 0xff, 0x01, 0x0e, 0xf0, 0x01, 0x62, 0x01, 0xd6, 0x00}
	data := []byte{0x10, 0x81, 0x0, 0x0, 0x05, 0xff, 0x01, 0x0e, 0xf0, 0x01, 0x62, 0x08, 0x80, 0x00, 0x82, 0x00, 0xd3, 0x00, 0xd4, 0x00, 0xd5, 0x00, 0xd6, 0x00, 0xd7, 0x00, 0x9f, 0x00}
	frame, err := ParseFrame(data)
	if err != nil {
		log.Print("Error:", err)
		return nil
	}
	log.Print(frame)
	return &frame
}

func createAirconGetFrame() *Frame {
	// Get
	//data := []byte{0x10, 0x81, 0x0, 0x0, 0x05, 0xff, 0x01, 0x01, 0x30, 0x01, 0x62, 0x02, 0x80, 0x00, 0x9f, 0x00}
	data := []byte{}

	ehd1 := []byte{0x10}
	ehd2 := []byte{0x81}
	tid := []byte{0x0, 0x0}

	edata := []byte{}
	seoj := []byte{0x05, 0xff, 0x01}
	deoj := []byte{0x01, 0x30, 0x01}
	esv := []byte{0x62}

	properties := []struct {
		epc []byte
		edt []byte
	}{
		{epc: []byte{0xbb}, edt: []byte{}},
		{epc: []byte{0xbe}, edt: []byte{}},
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

	frame, err := ParseFrame(data)
	if err != nil {
		log.Print("Error:", err)
		return nil
	}
	log.Print(frame)
	return &frame
}
