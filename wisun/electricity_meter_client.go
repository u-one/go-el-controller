package wisun

import (
	"bytes"
	"context"
	"fmt"
	"github.com/u-one/go-el-controller/echonetlite"
	"log"
	"time"
)

// ElectricityMeterClient is client for electricity meter
type ElectricityMeterClient struct {
	serialClient SerialClient
}

// NewElectricityMeterClient returns ElectricityMeterClient instance
func NewElectricityMeterClient(c SerialClient) *ElectricityMeterClient {
	return &ElectricityMeterClient{c}
}

// Close closes client
func (c ElectricityMeterClient) Close() {
	c.serialClient.Close()
}

// Version gets version
func (c ElectricityMeterClient) Version() error {
	err := c.serialClient.Send([]byte("SKVER\r\n"))
	_, err = c.serialClient.Recv()
	_, err = c.serialClient.Recv()
	_, err = c.serialClient.Recv()
	return err
}

// StartSequence starts PANA sequence
func (c ElectricityMeterClient) StartSequence(bRouteID, bRoutePW string) {
	c.serialClient.Send([]byte("SKVER\r\n"))
	c.serialClient.Recv()
	c.serialClient.Recv()
	c.serialClient.Recv()

	if len(bRouteID) == 0 {
		log.Fatal("set B-route ID")
	}
	if len(bRoutePW) == 0 {
		log.Fatal("set B-route password")
	}

	c.serialClient.Send([]byte("SKSETPWD C " + bRoutePW + "\r\n"))
	c.serialClient.Recv()
	c.serialClient.Recv()

	c.serialClient.Send([]byte("SKSETRBID " + bRouteID + "\r\n"))
	c.serialClient.Recv()
	c.serialClient.Recv()

	scan := func(duration int) bool {
		cmd := fmt.Sprintf("SKSCAN 2 FFFFFFFF %d 0 \r\n", duration)
		c.serialClient.Send([]byte(cmd))
		c.serialClient.Recv()
		c.serialClient.Recv()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		for {
			ch := make(chan error)
			var data []byte
			go func(data *[]byte) {
				res, err := c.serialClient.Recv()
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

	type epanDesc struct {
		Addr     string
		IPV6Addr string
		Channel  string
		PanID    string
	}

	recvEPANDesc := func() (epanDesc, error) {
		ed := epanDesc{}
		line, err := c.serialClient.Recv()
		if err == nil && bytes.HasPrefix(line, []byte("EPANDESC")) {
			line, err := c.serialClient.Recv() // Channel
			if err != nil {
				return epanDesc{}, fmt.Errorf("Failed to get Channel [%s]", err)
			}
			tokens := bytes.Split(line, []byte("Channel:"))
			ed.Channel = string(bytes.Trim(tokens[1], "\r\n"))
			c.serialClient.Recv()             // Channel Page: XX
			line, err = c.serialClient.Recv() // Pan ID: XXXX
			if err != nil {
				return epanDesc{}, fmt.Errorf("Failed to get Pan ID [%s]", err)
			}
			tokens = bytes.Split(line, []byte("Pan ID:"))
			ed.PanID = string(bytes.Trim(tokens[1], "\r\n"))
			line, err = c.serialClient.Recv() // Addr:XXXXXXXXXXXXXXXX
			if err != nil {
				return epanDesc{}, fmt.Errorf("Failed to get Addr [%s]", err)
			}
			tokens = bytes.Split(line, []byte("Addr:"))
			ed.Addr = string(bytes.Trim(tokens[1], "\r\n"))
			c.serialClient.Recv() // LQI:CA
			c.serialClient.Recv() // Side:X
			c.serialClient.Recv() // PairID:XXXXXXXX
		}
		return ed, err
	}
	ed, err := recvEPANDesc()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Received EPANDesc:%#v", ed)

	cmd := fmt.Sprintf("SKLL64 %s\r\n", ed.Addr)
	c.serialClient.Send([]byte(cmd))
	c.serialClient.Recv()
	line, err := c.serialClient.Recv()
	if err != nil {
		log.Fatal(err)
	}
	ed.IPV6Addr = string(bytes.Trim(line, "\r\n"))
	log.Printf("Translated address:%#v", ed)

	cmd = fmt.Sprintf("SKSREG S2 %s\r\n", ed.Channel)
	c.serialClient.Send([]byte(cmd))
	c.serialClient.Recv()
	c.serialClient.Recv()

	cmd = fmt.Sprintf("SKSREG S3 %s\r\n", ed.PanID)
	c.serialClient.Send([]byte(cmd))
	c.serialClient.Recv()
	c.serialClient.Recv()

	join := func(ed epanDesc) bool {
		cmd = fmt.Sprintf("SKJOIN %s\r\n", ed.IPV6Addr)
		c.serialClient.Send([]byte(cmd))
		c.serialClient.Recv()
		c.serialClient.Recv()

		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				log.Fatal(ctx.Err())
				return false
			default:
				res, err := c.serialClient.Recv()
				if err != nil {
					log.Println(err)
					if err.Error() == "serial: timeout" {
						continue
					}
					return false
				}

				if bytes.HasPrefix(res, []byte("EVENT 24")) {
					log.Println("found EVENT 24")
					return false
				}
				if bytes.HasPrefix(res, []byte("EVENT 25")) {
					log.Println("found EVENT 25")
					return true
				}
			}
		}
	}

	if !join(ed) {
		log.Fatal("Failed to join")
	}

	c.serialClient.Recv()
	c.serialClient.Recv()
	c.serialClient.Recv()
	c.serialClient.Recv()

}

func (c ElectricityMeterClient) GetCurrentPowerConsumption(ctx context.Context) (int ,error) {
	elframe := []byte{0x10, 0x81, 0x00, 0x01, 0x05, 0xff, 0x01, 0x02, 0x88, 0x01, 0x62, 0x01, 0xe7, 0x00}
	ipv6addr := "FE80:0000:0000:0000:021C:6400:030C:12A4"

	cmd := []byte(fmt.Sprintf("SKSENDTO 1 %s 0E1A 1 0 %04X ", ipv6addr, len(elframe)))
	cmd = append(cmd, elframe...)
	cmd = append(cmd, []byte("\r\n")...)
	c.serialClient.Send(cmd)

	for {
		select {
		case <-ctx.Done():
			log.Fatal(ctx.Err())
			return 0, ctx.Err()
		default:
			res, err := c.serialClient.Recv()
			if err != nil {
				log.Println(err)
				if err.Error() == "serial: timeout" {
					continue
				}
				return 0, err
			}

			// b'ERXUDP FE80:0000:0000:0000:021C:6400:030C:12A4 FE80:0000:0000:0000:021D:1291:0000:0574 0E1A 0E1A 001C6400030C12A4 1 0 0012 \x10\x81\x00\x01\x02\x88\x01\x05\xff\x01r\x01\xe7\x04\x00\x00\x01\xf8\r\n'
			if bytes.HasPrefix(res, []byte("ERXUDP")) {
				log.Println("found ERXUDP")
				// TODO: Trim and Append linebreak in Recv(), Send() method
				bytes.Trim(res, "\r\n")
				parseRXUDP(res)

				return 0,nil
			}
		}
	}
}

func parseRXUDP(line []byte) error {
	// b'ERXUDP FE80:0000:0000:0000:021C:6400:030C:12A4 FE80:0000:0000:0000:021D:1291:0000:0574 0E1A 0E1A 001C6400030C12A4 1 0 0012 \x10\x81\x00\x01\x02\x88\x01\x05\xff\x01r\x01\xe7\x04\x00\x00\x01\xf8\r\n'
	cols := bytes.Split(line, []byte{' '})
	if len(cols) < 10 {
		return fmt.Errorf("RXUDP invalid format")
	}
	elData := cols[9]
	elFrame, err := echonetlite.ParseFrame(elData)
	if err != nil {
		return fmt.Errorf("RXUDP invalid frame: %w", err)
	}
	elFrame.Print()
	return nil
}
