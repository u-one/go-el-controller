package wisun

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/goburrow/serial"
)

// BP35C2Emulator is ROHM BP35C2 emulator
type BP35C2Emulator struct {
	port     serial.Port
	reader   *bufio.Reader
	writer   *bufio.Writer
	curLine  []byte
	echoback bool
}

// NewBP35C2Emulator returns BP35C2Emulator instance
func NewBP35C2Emulator(addr string) *BP35C2Emulator {
	config := serial.Config{
		Address:  addr,
		BaudRate: 115200,
		DataBits: 8,
		StopBits: 1,
		Parity:   "N",
		Timeout:  30 * time.Second,
	}

	port, err := serial.Open(&config)
	if err != nil {
		log.Fatal("Faild to open serial:", err)
	}

	r := bufio.NewReaderSize(port, 4096)
	w := bufio.NewWriter(port)
	return &BP35C2Emulator{port: port, reader: r, writer: w, echoback: true}
}

// Close closees connection
func (e BP35C2Emulator) Close() {
	e.port.Close()
}

func (e BP35C2Emulator) flush() {
	e.writer.Flush()
}

func (e BP35C2Emulator) echoBack() {
	if e.echoback {
		e.writer.Write(e.curLine)
		e.writer.WriteString("\r\n")
		e.flush()
		fmt.Println("=>", string(e.curLine))
	}
}

func (e BP35C2Emulator) ok() {
	e.writer.WriteString("OK\r\n")
	e.flush()
	fmt.Println("=>OK")
}

func (e BP35C2Emulator) rxUDP(data []byte) {
	dlen := len(data)
	cmd := fmt.Sprintf("ERXUDP FE80:0000:0000:0000:021C:6400:030C:12A4 FF02:0000:0000:0000:0000:0000:0000:0001 0E1A 0E1A 001C6400030C12A4 1 0 %04x ", dlen)
	e.writer.WriteString(cmd)
	e.writer.Write(data)
	e.writer.WriteString("\r\n")
	e.flush()
	fmt.Println(cmd)
}

// Start starts to emulate
func (e *BP35C2Emulator) Start() {
	fmt.Println("Start")
	for {
		line, _, err := e.reader.ReadLine()
		if err != nil {
			fmt.Println("[Error]", err)
		}
		e.curLine = line
		fmt.Println("<=", string(line))
		if bytes.HasPrefix(line, []byte("SKVER")) {
			e.echoBack()
			e.writer.WriteString("EVER 1.0.0\r\n")
			e.flush()
			e.ok()
		} else if bytes.HasPrefix(line, []byte("SKSETPWD")) {
			e.echoBack()
			e.ok()
		} else if bytes.HasPrefix(line, []byte("SKSETRBID")) {
			e.echoBack()
			e.ok()
		} else if bytes.HasPrefix(line, []byte("SKSCAN")) {
			e.echoBack()
			e.ok()

			cmd := string(line)
			params := strings.Split(cmd, " ")
			if params[3] == "4" {
				e.writer.WriteString("EVENT 22 FE80:0000:0000:0000:021D:1290:1234:5678 0\r\n")
				e.flush()
			} else if params[3] == "5" {
				e.writer.WriteString("EVENT 20 FE80:0000:0000:0000:021D:1290:1234:5678 0\r\n")
				e.flush()

				e.writer.WriteString("EPANDESC\r\n")
				e.writer.WriteString(" Channel:21\r\n")
				e.writer.WriteString(" Channel Page:09\r\n")
				e.writer.WriteString(" Pan ID:8888\r\n")
				e.writer.WriteString(" Addr:12345678ABCDEF01\r\n")
				e.writer.WriteString(" LQI:E1\r\n")
				e.writer.WriteString(" Side:0\r\n")
				e.writer.WriteString(" PairID:AABBCCDD\r\n")
				e.flush()
			}

		} else if bytes.HasPrefix(line, []byte("SKLL64")) {
			e.echoBack()

			e.writer.WriteString("FE80:0000:0000:0000:021D:1290:1234:ABCD\r\n")
			e.flush()

		} else if bytes.HasPrefix(line, []byte("SKSREG")) {
			e.echoBack()
			e.ok()
		} else if bytes.HasPrefix(line, []byte("SKJOIN")) {
			e.echoBack()
			e.ok()

			e.writer.WriteString("EVENT 22 FE80:0000:0000:0000:021D:1290:1234:5678 0\r\n")
			e.flush()

			e.writer.WriteString("EVENT 21 FE80:0000:0000:0000:021D:1290:1234:5678 0 00\r\n")
			e.flush()
			e.writer.WriteString("ERXUDP FE80:0000:0000:0000:021D:1290:1234:ABCD FE80:0000:0000:0000:021D:1290:1234:5678 02CC 02CC 12345678ABCDEF01 0 0 0028 ")
			e.writer.Write([]byte{
				0x00, 0x00, // Reserved (16bit)
				0x00, 0x28, // Message Length (16bit)
				0xc0, 0x00, // Flags (16bit) [RSCAPIrrrrrrrrrr] R:Request, S:Start, C:Complete, A:re-Authentication, P:Ping, I:IP Reconfiguration, r:reserved, 11000000 00000000 -> Request&Start
				0x00, 0x02, // Message Type (16bit)
				0x1c, 0x2f, 0xf4, 0xb9, // Session Identifier (32bit)
				0x23, 0x84, 0x41, 0x58, // Sequence Number (32bit)
				// AVPs
				0x00, 0x06, // AVP Code (16bit)
				0x00, 0x00, // AVP Flags (16bit) [Vrrrrrrrrrrrrrrr] V:Vendor, r:reserved
				0x00, 0x04, // AVP Length (16bit), num of octets
				0x00, 0x00, // Reserved (16bit)
				// Vendor-ID (None. Present if V bit in AVP Flags is set)
				0x00, 0x00, 0x00, 0x05, // Value
				0x00, 0x03, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0c})

			e.writer.WriteString("\r\n")
			e.flush()

			e.writer.WriteString("EVENT 21 FE80:0000:0000:0000:021D:1290:1234:5678 0 00\r\n")
			e.flush()
			e.writer.WriteString("ERXUDP FE80:0000:0000:0000:021D:1290:1234:ABCD FE80:0000:0000:0000:021D:1290:1234:5678 02CC 02CC 12345678ABCDEF01 0 0 0068 ")
			e.writer.Write([]byte{
				0x00, 0x00, 0x00, 0x68, 0x80, 0x00, 0x00, 0x02, 0x1c, 0x2f, 0xf4, 0xb9, 0x23, 0x84, 0x41, 0x59, // Flags: 10000000 00000000 -> Request
				0x00, 0x05, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x4b, 0x5a, 0xd1, 0x90, 0x99, 0x8e, 0xe9, 0x2e,
				0xfd, 0x8a, 0x95, 0xfc, 0x29, 0x38, 0x9f, 0xa2, 0x00, 0x02, 0x00, 0x00, 0x00, 0x38, 0x00, 0x00,
				0x01, 0x48, 0x00, 0x38, 0x2f, 0x00, 0x98, 0x3b, 0x5c, 0x1b, 0x72, 0x33, 0x26, 0xe7, 0xbe, 0x2b,
				0x4c, 0x07, 0xe8, 0x09, 0x0c, 0xf1, 0x53, 0x4d, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x39, 0x39,
				0x30, 0x32, 0x31, 0x31, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
				0x30, 0x31, 0x31, 0x32, 0x43, 0x45, 0x36, 0x37})
			e.writer.WriteString("\r\n")
			e.flush()

			e.writer.WriteString("EVENT 21 FE80:0000:0000:0000:021D:1290:1234:5678 0 00\r\n")
			e.flush()
			e.writer.WriteString("ERXUDP FE80:0000:0000:0000:021D:1290:1234:ABCD FE80:0000:0000:0000:021D:1290:1234:5678 02CC 02CC 12345678ABCDEF01 0 0 0054 ")
			e.writer.Write([]byte{
				0x00, 0x00, 0x00, 0x54, 0x80, 0x00, 0x00, 0x02, 0x1c, 0x2f, 0xf4, 0xb9, 0x23, 0x84, 0x41, 0x5a, // Flags: 10000000 00000000 -> Request
				0x00, 0x02, 0x00, 0x00, 0x00, 0x3b, 0x00, 0x00, 0x01, 0x49, 0x00, 0x3b, 0x2f, 0x80, 0x98, 0x3b,
				0x5c, 0x1b, 0x72, 0x33, 0x26, 0xe7, 0xbe, 0x2b, 0x4c, 0x07, 0xe8, 0x09, 0x0c, 0xf1, 0x19, 0x88,
				0x78, 0x8f, 0x48, 0x5e, 0x69, 0x94, 0xed, 0x46, 0xf0, 0xf5, 0x36, 0x1e, 0x9a, 0xb7, 0x00, 0x00,
				0x00, 0x00, 0x73, 0x97, 0x16, 0xc5, 0x80, 0xad, 0x49, 0x62, 0x17, 0xb7, 0x68, 0x8a, 0xe4, 0x6e,
				0xff, 0xaf, 0x7e, 0x00})
			e.writer.WriteString("\r\n")
			e.flush()

			e.writer.WriteString("EVENT 21 FE80:0000:0000:0000:021D:1290:1234:5678 0 00\r\n")
			e.flush()
			e.writer.WriteString("ERXUDP FE80:0000:0000:0000:021D:1290:1234:ABCD FE80:0000:0000:0000:021D:1290:1234:5678 02CC 02CC 12345678ABCDEF01 0 0 0058 ")
			e.writer.Write([]byte{
				0x00, 0x00, 0x00, 0x58, 0xa0, 0x00, 0x00, 0x02, 0x1c, 0x2f, 0xf4, 0xb9, 0x23, 0x84, 0x41, 0x5b, // Flags: 10100000 00000000 -> Request&Complete
				0x00, 0x07, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00,
				0x00, 0x04, 0x00, 0x00, 0x03, 0x49, 0x00, 0x04, 0x00, 0x04, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00,
				0x00, 0x00, 0x12, 0x10, 0x00, 0x08, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x01, 0x51, 0x80,
				0x00, 0x01, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x10, 0x8e, 0x79, 0x57, 0xb9, 0xb2, 0x6c, 0x04,
				0x34, 0x5d, 0x70, 0x25, 0xc2, 0x4a, 0x24, 0x72})
			e.writer.WriteString("\r\n")
			e.flush()

			e.writer.WriteString("EVENT 21 FE80:0000:0000:0000:021D:1290:1234:5678 0 00\r\n")
			e.flush()
			e.writer.WriteString("EVENT 25 FE80:0000:0000:0000:021D:1290:1234:5678 0\r\n")
			e.flush()

		} else if bytes.HasPrefix(line, []byte("SKSENDTO")) {
			e.echoBack()

			e.writer.WriteString("EVENT 21 FE80:0000:0000:0000:021D:1290:1234:5678 0 00\r\n")
			e.flush()

			e.ok()
			e.rxUDP([]byte{0x10, 0x81, 0x00, 0x01, 0x02, 0x88, 0x01, 0x05, 0xff, 0x01, 'r', 0x01, 0xe7, 0x04, 0x00, 0x00, 0x01, 0xf8})
		} else if bytes.HasPrefix(line, []byte("SKTERM")) {
			e.echoBack()
			e.ok()
		}

	}
}
