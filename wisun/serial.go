package wisun

import (
	"bufio"
	"log"
	"time"

	"github.com/goburrow/serial"
)

//go:generate mockgen -source serial.go -destination serial_mock.go -package wisun

type Serial interface {
	Send([]byte) error
	Recv() ([]byte, error)
	Close()
}

type SerialImpl struct {
	port   serial.Port
	reader *bufio.Reader
}

func NewSerialImpl() *SerialImpl {
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

	return &SerialImpl{port: port, reader: reader}
}

func (s SerialImpl) Send(in []byte) error {
	_, err := s.port.Write(in)
	return err
}

func (s SerialImpl) Recv() ([]byte, error) {
	line, _, err := s.reader.ReadLine()
	return line, err
}

func (s SerialImpl) Close() {
	s.port.Close()
}
