package wisun

import (
	"bufio"
	"log"
	"time"

	"github.com/goburrow/serial"
)

//go:generate mockgen -source wisun.go -destination wisun_mock.go -package wisun

// SerialClient is serial interface
type SerialClient interface {
	Send(in []byte) error
	Recv() ([]byte, error)
	Close()
}

// BP35C2Client is client for ROHM BP35C2
type BP35C2Client struct {
	sendSeq int
	readSeq int
	port    serial.Port
	reader  *bufio.Reader
}

// NewBP35C2Client returns BP35C2Client instance
func NewBP35C2Client() *BP35C2Client {
	config := serial.Config{
		Address:  "/dev/ttyUSB0",
		BaudRate: 115200,
		DataBits: 8,
		StopBits: 1,
		Parity:   "N",
		Timeout:  1 * time.Second,
	}

	port, err := serial.Open(&config)
	if err != nil {
		log.Fatal("Faild to open serial:", err)
	}

	reader := bufio.NewReaderSize(port, 4096)

	return &BP35C2Client{port: port, reader: reader}
}

// Close closees connection
func (c BP35C2Client) Close() {
	c.port.Close()
}

// Send sends serial command
func (c *BP35C2Client) Send(in []byte) error {
	c.sendSeq++
	log.Printf("Send[%d]:%s", c.sendSeq, string(in))
	var err error
	if _, err = c.port.Write(in); err != nil {
		log.Fatal(err)
	}
	return err
}

// Recv receives serial response by line
func (c *BP35C2Client) Recv() ([]byte, error) {
	line, _, err := c.reader.ReadLine()
	c.readSeq++
	if err != nil {
		//log.Fatal(err)
		log.Println(err)
		return []byte{}, err
	}
	log.Printf("Read[%d]:%s", c.readSeq, string(line))
	return line, err
}
