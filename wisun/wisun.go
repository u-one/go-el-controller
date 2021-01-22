package wisun

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/goburrow/serial"
)

//go:generate mockgen -source wisun.go -destination wisun_mock.go -package wisun

// Client is wisun client
type Client interface {
	Version() error
	SetBRoutePassword(password string) error
	SetBRouteID(id string) error
	Scan() (PanDesc, error)
	LL64(addr string) (string, error)
	SRegS2(channel string) error
	SRegS3(panID string) error
	Join(desc PanDesc) (bool, error)
	Close()
	SendTo(ipv6addr string, data []byte) ([]byte, error)
}

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

// PanDesc is...
type PanDesc struct {
	Addr     string
	IPV6Addr string
	Channel  string
	PanID    string
}

// BP35C2Client is client for ROHM BP35C2
type BP35C2Client struct {
	sendSeq int
	readSeq int
	serial  Serial
}

// NewBP35C2Client returns BP35C2Client instance
func NewBP35C2Client() *BP35C2Client {
	s := NewSerialImpl()
	return &BP35C2Client{serial: s}
}

// Close closees connection
func (c BP35C2Client) Close() {
	c.serial.Close()
}

// Version is ..
func (c BP35C2Client) Version() error {
	err := c.send([]byte("SKVER\r\n"))
	_, err = c.recv()
	_, err = c.recv()
	_, err = c.recv()
	return err
}

// SetBRoutePassword is..
func (c BP35C2Client) SetBRoutePassword(password string) error {
	if len(password) == 0 {
		return fmt.Errorf("set B-route password")
	}

	c.send([]byte("SKSETPWD C " + password + "\r\n"))
	c.recv()
	c.recv()
	return nil
}

// SetBRouteID  is ..
func (c BP35C2Client) SetBRouteID(id string) error {
	if len(id) == 0 {
		return fmt.Errorf("set B-route ID")
	}

	c.send([]byte("SKSETRBID " + id + "\r\n"))
	c.recv()
	c.recv()
	return nil
}

// Scan is ..
func (c BP35C2Client) Scan() (PanDesc, error) {
	scan := func(duration int) bool {
		cmd := fmt.Sprintf("SKSCAN 2 FFFFFFFF %d 0 \r\n", duration)
		c.send([]byte(cmd))
		c.recv()
		c.recv()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		for {
			ch := make(chan error)
			var data []byte
			go func(data *[]byte) {
				res, err := c.recv()
				if err != nil {
					log.Println(err)
					ch <- err
				}

				if bytes.HasPrefix(res, []byte("EVENT 22")) {
					log.Println("found EVENT 22")
					ch <- nil
				}
				if bytes.HasPrefix(res, []byte("EVENT 20")) {
					log.Println("found EVENT 20")
					*data = res
					ch <- nil
				}
			}(&data)

			select {
			case err := <-ch:
				if err == nil {
					if len(data) == 0 {
						return false
					}
					return true
				}
			case <-ctx.Done():
				log.Fatal(ctx.Err())
				return false
			}
		}
	}

	duration := 4
	for {
		if duration > 8 {
			log.Println("duration limit(8) exceeds")
			break
		}

		found := scan(duration)
		if found {
			break
		}
		duration = duration + 1
	}

	recvEPANDesc := func() (PanDesc, error) {
		ed := PanDesc{}
		line, err := c.recv()
		if err == nil && bytes.HasPrefix(line, []byte("EPANDESC")) {
			line, err := c.recv() // Channel
			if err != nil {
				return PanDesc{}, fmt.Errorf("Failed to get Channel [%s]", err)
			}
			tokens := bytes.Split(line, []byte("Channel:"))
			ed.Channel = string(bytes.Trim(tokens[1], "\r\n"))
			c.recv()             // Channel Page: XX
			line, err = c.recv() // Pan ID: XXXX
			if err != nil {
				return PanDesc{}, fmt.Errorf("Failed to get Pan ID [%s]", err)
			}
			tokens = bytes.Split(line, []byte("Pan ID:"))
			ed.PanID = string(bytes.Trim(tokens[1], "\r\n"))
			line, err = c.recv() // Addr:XXXXXXXXXXXXXXXX
			if err != nil {
				return PanDesc{}, fmt.Errorf("Failed to get Addr [%s]", err)
			}
			tokens = bytes.Split(line, []byte("Addr:"))
			ed.Addr = string(bytes.Trim(tokens[1], "\r\n"))
			c.recv() // LQI:CA
			c.recv() // Side:X
			c.recv() // PairID:XXXXXXXX
		}
		return ed, err
	}
	ed, err := recvEPANDesc()
	log.Printf("Received EPANDesc:%#v", ed)
	return ed, err
}

// LL64 is .
func (c BP35C2Client) LL64(addr string) (string, error) {
	cmd := fmt.Sprintf("SKLL64 %s\r\n", addr)
	c.send([]byte(cmd))
	c.recv()
	line, err := c.recv()
	if err != nil {
		return "", err
	}
	ipV6Addr := string(bytes.Trim(line, "\r\n"))
	log.Printf("Translated address:%#v", ipV6Addr)
	return ipV6Addr, nil
}

// SRegS2 is.
func (c BP35C2Client) SRegS2(channel string) error {
	cmd := fmt.Sprintf("SKSREG S2 %s\r\n", channel)
	c.send([]byte(cmd))
	c.recv()
	c.recv()
	return nil
}

// SRegS3 is ..
func (c BP35C2Client) SRegS3(panID string) error {
	cmd := fmt.Sprintf("SKSREG S3 %s\r\n", panID)
	c.send([]byte(cmd))
	c.recv()
	c.recv()
	return nil
}

// Join is ..
func (c BP35C2Client) Join(desc PanDesc) (bool, error) {
	cmd := fmt.Sprintf("SKJOIN %s\r\n", desc.IPV6Addr)
	c.send([]byte(cmd))
	c.recv()
	c.recv()

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			log.Fatal()
			return false, fmt.Errorf("timeout:%w", ctx.Err())
		default:
			res, err := c.recv()
			if err != nil {
				log.Println(err)
				if err.Error() == "serial: timeout" {
					continue
				}
				return false, fmt.Errorf("join failed: %w", err)
			}

			if bytes.HasPrefix(res, []byte("EVENT 24")) {
				log.Println("found EVENT 24")
				return false, nil
			}
			if bytes.HasPrefix(res, []byte("EVENT 25")) {
				log.Println("found EVENT 25")
				return true, nil
			}
		}
	}

}

// SendTo is...
func (c *BP35C2Client) SendTo(ipv6Addr string, data []byte) ([]byte, error) {
	cmd := []byte(fmt.Sprintf("SKSENDTO 1 %s 0E1A 1 0 %04X ", ipv6Addr, len(data)))
	cmd = append(cmd, data...)
	cmd = append(cmd, []byte("\r\n")...)
	c.send(cmd)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			log.Fatal(ctx.Err())
			return nil, ctx.Err()
		default:
			res, err := c.recv()
			if err != nil {
				log.Println(err)
				if err.Error() == "serial: timeout" {
					continue
				}
				return nil, err
			}

			// b'ERXUDP FE80:0000:0000:0000:021C:6400:030C:12A4 FE80:0000:0000:0000:021D:1291:0000:0574 0E1A 0E1A 001C6400030C12A4 1 0 0012 \x10\x81\x00\x01\x02\x88\x01\x05\xff\x01r\x01\xe7\x04\x00\x00\x01\xf8\r\n'
			if bytes.HasPrefix(res, []byte("ERXUDP")) {
				log.Println("found ERXUDP")
				// TODO: Trim and Append linebreak in recv(), Send() method
				bytes.Trim(res, "\r\n")
				rdata, err := parseRXUDP(res)
				return rdata, err
			}
		}
	}
}

func parseRXUDP(line []byte) ([]byte, error) {
	// b'ERXUDP FE80:0000:0000:0000:021C:6400:030C:12A4 FE80:0000:0000:0000:021D:1291:0000:0574 0E1A 0E1A 001C6400030C12A4 1 0 0012 \x10\x81\x00\x01\x02\x88\x01\x05\xff\x01r\x01\xe7\x04\x00\x00\x01\xf8\r\n'
	cols := bytes.Split(line, []byte{' '})
	if len(cols) < 10 {
		return nil, fmt.Errorf("RXUDP invalid format")
	}
	return cols[9], nil
}

// Send sends serial command
func (c *BP35C2Client) send(in []byte) error {
	c.sendSeq++
	log.Printf("Send[%d]:%s", c.sendSeq, string(in))
	var err error
	if err = c.serial.Send(in); err != nil {
		log.Fatal(err)
	}
	return err
}

// recv receives serial response by line
func (c *BP35C2Client) recv() ([]byte, error) {
	line, err := c.serial.Recv()
	c.readSeq++
	if err != nil {
		return []byte{}, err
	}
	log.Printf("Read[%d]:%s", c.readSeq, string(line))
	return line, err
}
