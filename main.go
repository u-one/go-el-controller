package main

import (
	"context"
	"encoding/csv"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

var (
	// MulticastIP is Echonet-Lite multicast address
	MulticastIP = "224.0.23.0"
	// Port is Echonet-Lite receive port
	Port = ":3610"
)
var (
	// ClassMap is a map with ClassCode as key and PropertyDefs as value
	ClassMap *map[ClassCode]PropertyDefs
)

// Data represents binary data
type Data []byte

// String returns hex representation for the data
func (d Data) String() string {
	return hex.EncodeToString(d)
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

// NewFrame returns Frame
func NewFrame(data []byte) (Frame, error) {
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

	epcOffset := 8
	EPC := EDATA[epcOffset : epcOffset+1]
	PDC := EDATA[epcOffset+1 : epcOffset+2]
	propertyValueLen := int(PDC[0])
	EDT := EDATA[epcOffset+2 : epcOffset+2+propertyValueLen]

	log.Println("EHD:", EHD)
	log.Println("TID:", TID)
	log.Println("SEOJ:", SEOJ, " (Source Echonet lite object)")
	log.Println("DEOJ:", DEOJ, " (Dest. Echonet lite object)")
	log.Println("ESV:", ESV, " (Echonet Lite Service)")
	log.Println("OPC:", OPC, " (Num of properties)")

	props := (*ClassMap)[NewClassCode(SEOJ[:2])]
	desc := ""
	if props != nil {
		desc = props[EPCCode(EPC[0])]
	}

	log.Println("EPC:", EPC, " (Echonet lite property) ", desc)
	log.Println("PDC:", PDC, " (Length (bytes) of EDT)")
	log.Println("EDT:", EDT, " (Property data)")

	return Frame{Data: frame, EHD: EHD, TID: TID, EData: EDATA, SEOJ: SEOJ, DEOJ: DEOJ, ESV: ESV, OPC: OPC, Properties: []Property{}}, nil
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

// Property represents Echonet-Lite property
type Property Data

func read(ctx context.Context) chan bool {
	ch := make(chan bool)
	go func() {
		defer close(ch)
		log.Println("Start to listen multicast udp ", MulticastIP, Port)
		address, err := net.ResolveUDPAddr("udp", MulticastIP+Port)
		if err != nil {
			log.Println("Error: ", err)
			return
		}
		conn, err := net.ListenMulticastUDP("udp", nil, address)
		if err != nil {
			log.Println("Error:", err)
			return
		}
		defer conn.Close()
		buffer := make([]byte, 1500)

		for {
			fmt.Printf(".")
			conn.SetDeadline(time.Now().Add(1 * time.Second))
			length, remoteAddress, err := conn.ReadFromUDP(buffer)
			if err != nil {
				log.Println("Error: ", err)
			}
			if length > 0 {
				fmt.Println()
				frame, err := NewFrame(buffer[:length])
				if err != nil {
					log.Println(err)
					continue
				}
				log.Printf("[%v] %v\n", remoteAddress, frame)
			}
			select {
			case <-ctx.Done():
				log.Println("ctx.Done")
				ch <- true
				return
			default:
				//log.Println("recv: ", length)
			}
		}
	}()
	return ch
}

// ClassCode is Echonet-Lite Class information
type ClassCode struct {
	ClassGroupCode byte
	ClassCode      byte
}

// NewClassCode returns new instance of ClassCode
func NewClassCode(code []byte) ClassCode {
	log.Println("NewClassCode:", code)
	return ClassCode{
		ClassGroupCode: code[0],
		ClassCode:      code[1],
	}
}

// EPCCode is EPC code
type EPCCode byte

// PropertyDefs is a map EPCCode as key detail string as value
type PropertyDefs map[EPCCode]string

func load() *map[ClassCode]PropertyDefs {
	csvPath := "../../SonyCSL/ECHONETLite-ObjectDatabase/data/csv/ja"
	files, err := ioutil.ReadDir(csvPath)
	if err != nil {
		log.Fatal(err)
	}

	classMap := make(map[ClassCode]PropertyDefs)

	for _, file := range files {
		func() {
			log.Println(csvPath, file.Name())
			name := strings.Split(file.Name(), ".")[0]

			if !strings.HasPrefix(name, "0x") {
				log.Println("Not property file: ", name)
				return
			}

			decodedClassCodes, err := hex.DecodeString(strings.TrimPrefix(name, "0x"))
			log.Println(decodedClassCodes)

			properties := make(PropertyDefs, 0)

			path := csvPath + "/" + file.Name()
			f, err := os.Open(path)
			defer f.Close()
			if err != nil {
				log.Fatal(err)
				return
			}
			var epcBegan = false

			r := csv.NewReader(f)
			for {
				record, err := r.Read()
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Fatal(err)
					continue
				}
				if record[0] == "EPC" {
					epcBegan = true
					continue
				}
				if !epcBegan {
					continue
				}
				log.Println(record[0])
				if !strings.HasPrefix(record[0], "0x") {
					continue
				}
				if len(record[0]) == 0 {
					continue
				}
				d, err := hex.DecodeString(strings.TrimPrefix(record[0], "0x"))
				if err != nil {
					log.Println("failed to decode:", record[0])
					continue
				}
				properties[EPCCode(d[0])] = record[1]
			}
			c := Data(decodedClassCodes)
			//log.Println("Decoded class code", c)
			classMap[NewClassCode(c)] = properties
		}()
	}
	return &classMap
}

func write() {
	conn, err := net.Dial("udp", MulticastIP+Port)
	if err != nil {
		log.Println("Write conn error:", err)
		return
	}
	defer conn.Close()

	data := []byte{0x10, 0x81, 0x0, 0x0, 0x0e, 0xf0, 0x01, 0x0e, 0xf0, 0x01, 0x73, 0x01, 0xd5, 0x04, 0x01, 0x05, 0xff, 0x01}
	//data := []byte{0x10, 0x81, 0x0, 0x0, 0x05, 0xff, 0x01, 0x0e, 0xf0, 0x01, 0x63, 0x01, 0xd5, 0x00}

	length, err := conn.Write(data)
	if err != nil {
		log.Println("Write error: ", err)
	}
	log.Println("written:", length)
}

func start() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	done := read(ctx)

	write()

	log.Println("wait for read done")
	res := <-done
	log.Println("read finish: ", res)
}

func main() {
	flag.Parse()

	ClassMap = load()
	start()
}
