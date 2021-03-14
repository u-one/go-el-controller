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
			e.writer.WriteString("ERXUDP FE80:0000:0000:0000:021D:1290:1234:ABCD FE80:0000:0000:0000:021D:1290:1234:5678 02CC 02CC 12345678ABCDEF01 0 0 0028\r\n")
			e.flush()

			e.writer.WriteString("EVENT 21 FE80:0000:0000:0000:021D:1290:1234:5678 0 00\r\n")
			e.flush()
			e.writer.WriteString("ERXUDP FE80:0000:0000:0000:021D:1290:1234:ABCD FE80:0000:0000:0000:021D:1290:1234:5678 02CC 02CC 12345678ABCDEF01 0 0 0068\r\n")
			e.flush()

			e.writer.WriteString("EVENT 21 FE80:0000:0000:0000:021D:1290:1234:5678 0 00\r\n")
			e.flush()
			e.writer.WriteString("ERXUDP FE80:0000:0000:0000:021D:1290:1234:ABCD FE80:0000:0000:0000:021D:1290:1234:5678 02CC 02CC 12345678ABCDEF01 0 0 0054\r\n")
			e.flush()

			e.writer.WriteString("EVENT 21 FE80:0000:0000:0000:021D:1290:1234:5678 0 00\r\n")
			e.flush()
			e.writer.WriteString("ERXUDP FE80:0000:0000:0000:021D:1290:1234:ABCD FE80:0000:0000:0000:021D:1290:1234:5678 02CC 02CC 12345678ABCDEF01 0 0 0058\r\n")
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
