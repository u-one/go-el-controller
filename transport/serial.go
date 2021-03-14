package transport

import (
	"bufio"
	"log"
	"time"

	"github.com/goburrow/serial"
)

//go:generate mockgen -source serial.go -destination serial_mock.go -package transport

// Serial is the interface that communicates data through serial port
type Serial interface {
	Send([]byte) error     // sends data
	Recv() ([]byte, error) // receives data and returns them
	Close()                // closes active serial connection
}

// SerialImpl is a concrete implementation of Serial interface
type SerialImpl struct {
	port   serial.Port
	reader *bufio.Reader
}

// NewSerialImpl opens default serial connection and returns SerialImpl
func NewSerialImpl(addr string) *SerialImpl {
	config := serial.Config{
		Address:  addr,
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

	return &SerialImpl{port: port, reader: reader}
}

// Send sends data
func (s SerialImpl) Send(in []byte) error {
	_, err := s.port.Write(in)
	return err
}

// Recv receives data
func (s SerialImpl) Recv() ([]byte, error) {
	line, _, err := s.reader.ReadLine()
	return line, err
}

// Close closes active connection
func (s SerialImpl) Close() {
	s.port.Close()
}
