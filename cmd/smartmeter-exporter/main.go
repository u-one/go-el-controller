package main

import (
	//	"io"
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"

	//	"os"
	"strings"
	"time"

	"github.com/goburrow/serial"
)

var bRouteID = flag.String("brouteid", "", "B-route ID")
var bRoutePW = flag.String("broutepw", "", "B-route password")

func main() {
	flag.Parse()

	wisunClient := NewBP35C2Client()
	c := NewElectricityMeterClient(wisunClient)
	defer c.Close()

	c.StartSequence(*bRouteID, *bRoutePW)
}

type SerialClient interface {
	Send(cmd string) error
	Recv() (string, error)
	Close()
}

type BP35C2Client struct {
	sendSeq int
	readSeq int
	port serial.Port 
	reader *bufio.Reader
}

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

func (c BP35C2Client) Close() {
	c.port.Close()
}

func (c *BP35C2Client) Send(cmd string) error {
	c.sendSeq += 1
	log.Printf("Send[%d]:%s", c.sendSeq, string(cmd))
	var err error
	if _, err = c.port.Write([]byte(cmd)); err != nil {
		log.Fatal(err)
	}
	return err
}

func (c *BP35C2Client) Recv() (string, error) {
	line, _, err := c.reader.ReadLine()
	c.readSeq += 1
	if err != nil {
		//log.Fatal(err)
		log.Println(err)
		return "", err
	}
	log.Printf("Read[%d]:%s", c.readSeq, string(line))
	return string(line), err
}

type ElectricityMeterClient struct {
	serialClient SerialClient
}

func NewElectricityMeterClient(c SerialClient) *ElectricityMeterClient {
	return &ElectricityMeterClient{c}
}

func (c ElectricityMeterClient) Close() {
	c.serialClient.Close() 
}

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

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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


