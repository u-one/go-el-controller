package wisun

import (
	"context"
	"fmt"
	"log"
	"strings"
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
	err := c.serialClient.Send("SKVER\r\n")
	_, err = c.serialClient.Recv()
	_, err = c.serialClient.Recv()
	_, err = c.serialClient.Recv()
	return err
}

// StartSequence starts PANA sequence
func (c ElectricityMeterClient) StartSequence(bRouteID, bRoutePW string) {
	c.serialClient.Send("SKVER\r\n")
	c.serialClient.Recv()
	c.serialClient.Recv()
	c.serialClient.Recv()

	if len(bRouteID) == 0 {
		log.Fatal("set B-route ID")
	}
	if len(bRoutePW) == 0 {
		log.Fatal("set B-route password")
	}

	c.serialClient.Send("SKSETPWD C " + bRoutePW + "\r\n")
	c.serialClient.Recv()
	c.serialClient.Recv()

	c.serialClient.Send("SKSETRBID " + bRouteID + "\r\n")
	c.serialClient.Recv()
	c.serialClient.Recv()

	scan := func(duration int) bool {
		cmd := fmt.Sprintf("SKSCAN 2 FFFFFFFF %d 0 \r\n", duration)
		c.serialClient.Send(cmd)
		c.serialClient.Recv()
		c.serialClient.Recv()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		for {
			ch := make(chan error)
			data := ""
			go func(data *string) {
				res, err := c.serialClient.Recv()
				if err != nil {
					log.Println(err)
					ch <- err
				}

				if strings.HasPrefix(res, "EVENT 22") {
					log.Println("found EVENT 22")
					ch <- nil
				}
				if strings.HasPrefix(res, "EVENT 20") {
					log.Println("found EVENT 20")
					*data = res
					ch <- nil
				}
			}(&data)

			select {
			case err := <-ch:
				if err == nil {
					if data == "" {
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
		if err == nil && strings.HasPrefix(line, "EPANDESC") {
			line, err := c.serialClient.Recv() // Channel
			if err != nil {
				return epanDesc{}, fmt.Errorf("Failed to get Channel [%s]", err)
			}
			tokens := strings.Split(line, "Channel:")
			ed.Channel = strings.Trim(tokens[1], "\r\n")
			c.serialClient.Recv()             // Channel Page: XX
			line, err = c.serialClient.Recv() // Pan ID: XXXX
			if err != nil {
				return epanDesc{}, fmt.Errorf("Failed to get Pan ID [%s]", err)
			}
			tokens = strings.Split(line, "Pan ID:")
			ed.PanID = strings.Trim(tokens[1], "\r\n")
			line, err = c.serialClient.Recv() // Addr:XXXXXXXXXXXXXXXX
			if err != nil {
				return epanDesc{}, fmt.Errorf("Failed to get Addr [%s]", err)
			}
			tokens = strings.Split(line, "Addr:")
			ed.Addr = strings.Trim(tokens[1], "\r\n")
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
	c.serialClient.Send(cmd)
	c.serialClient.Recv()
	line, err := c.serialClient.Recv()
	if err != nil {
		log.Fatal(err)
	}
	ed.IPV6Addr = strings.Trim(line, "\r\n")
	log.Printf("Translated address:%#v", ed)

	cmd = fmt.Sprintf("SKSREG S2 %s\r\n", ed.Channel)
	c.serialClient.Send(cmd)
	c.serialClient.Recv()
	c.serialClient.Recv()

	cmd = fmt.Sprintf("SKSREG S3 %s\r\n", ed.PanID)
	c.serialClient.Send(cmd)
	c.serialClient.Recv()
	c.serialClient.Recv()

	join := func(ed epanDesc) bool {
		cmd = fmt.Sprintf("SKJOIN %s\r\n", ed.IPV6Addr)
		c.serialClient.Send(cmd)
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

				if strings.HasPrefix(res, "EVENT 24") {
					log.Println("found EVENT 24")
					return false
				}
				if strings.HasPrefix(res, "EVENT 25") {
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
