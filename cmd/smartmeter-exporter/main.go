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
	defer port.Close()

	sendSeq := 0
	send := func(cmd string) error {
		sendSeq += 1
		log.Printf("Send[%d]:%s", sendSeq, string(cmd))
		if _, err = port.Write([]byte(cmd)); err != nil {
			log.Fatal(err)
		}
		return err
	}

	reader := bufio.NewReaderSize(port, 4096)

	readSeq := 0
	read := func() (string, error) {
		line, _, err := reader.ReadLine()
		readSeq += 1
		if err != nil {
			//log.Fatal(err)
			log.Println(err)
			return "", err
		}
		log.Printf("Read[%d]:%s", readSeq, string(line))
		return string(line), err
	}

	send("SKVER\r\n")
	read()
	read()
	read()

	if len(*bRouteID) == 0 {
		log.Fatal("set B-route ID")
	}
	if len(*bRoutePW) == 0 {
		log.Fatal("set B-route password")
	}

	send("SKSETPWD C " + *bRoutePW + "\r\n")
	read()
	read()

	send("SKSETRBID " + *bRouteID + "\r\n")
	read()
	read()

	scan := func(duration int) bool {
		cmd := fmt.Sprintf("SKSCAN 2 FFFFFFFF %d 0 \r\n", duration)
		send(cmd)
		read()
		read()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		for {
			ch := make(chan error)
			data := ""
			go func(data *string) {
				res, err := read()
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
		line, err := read()
		if err == nil && strings.HasPrefix(line, "EPANDESC") {
			line, err := read() // Channel
			if err != nil {
				return epanDesc{}, fmt.Errorf("Failed to get Channel [%s]", err)
			}
			tokens := strings.Split(line, "Channel:")
			ed.Channel = strings.Trim(tokens[1], "\r\n")
			read()             // Channel Page: XX
			line, err = read() // Pan ID: XXXX
			if err != nil {
				return epanDesc{}, fmt.Errorf("Failed to get Pan ID [%s]", err)
			}
			tokens = strings.Split(line, "Pan ID:")
			ed.PanID = strings.Trim(tokens[1], "\r\n")
			line, err = read() // Addr:XXXXXXXXXXXXXXXX
			if err != nil {
				return epanDesc{}, fmt.Errorf("Failed to get Addr [%s]", err)
			}
			tokens = strings.Split(line, "Addr:")
			ed.Addr = strings.Trim(tokens[1], "\r\n")
			read() // LQI:CA
			read() // Side:X
			read() // PairID:XXXXXXXX
		}
		return ed, err
	}
	ed, err := recvEPANDesc()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Received EPANDesc:%#v", ed)

	cmd := fmt.Sprintf("SKLL64 %s\r\n", ed.Addr)
	send(cmd)
	read()
	line, err := read()
	if err != nil {
		log.Fatal(err)
	}
	ed.IPV6Addr = strings.Trim(line, "\r\n")
	log.Printf("Translated address:%#v", ed)

	cmd = fmt.Sprintf("SKSREG S2 %s\r\n", ed.Channel)
	send(cmd)
	read()
	read()

	cmd = fmt.Sprintf("SKSREG S3 %s\r\n", ed.PanID)
	send(cmd)
	read()
	read()

	join := func(ed epanDesc) bool {
		cmd = fmt.Sprintf("SKJOIN %s\r\n", ed.IPV6Addr)
		send(cmd)
		read()
		read()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				log.Fatal(ctx.Err())
				return false
			default:
				res, err := read()
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

	read()
	read()
	read()
	read()

}
