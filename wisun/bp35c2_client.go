package wisun

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/u-one/go-el-controller/transport"
)

// BP35C2Client is client for ROHM BP35C2
type BP35C2Client struct {
	sendSeq int
	readSeq int
	serial  transport.Serial
	panDesc PanDesc
	joined  bool
}

// PanDesc is...
type PanDesc struct {
	Addr     string
	IPV6Addr string
	Channel  string
	PanID    string
}

// NewBP35C2Client returns BP35C2Client instance
func NewBP35C2Client(portaddr string) *BP35C2Client {
	fmt.Println("NewBP35C2Client: ", portaddr)
	s := transport.NewSerialImpl(portaddr)
	return &BP35C2Client{serial: s}
}

// Close closees connection
func (c *BP35C2Client) Close() {
	if c.joined {
		c.Term()
	}
	c.serial.Close()
}

func stringWithBinary(data []byte) string {
	// For debug
	var b strings.Builder
	data = bytes.TrimSuffix(data, []byte{'\r', '\n'})
	tokens := bytes.Split(data, []byte{' '})
	for i, token := range tokens {
		binary := false
		for _, r := range string(token) {
			if r == '\r' || r == '\n' {
				continue
			}
			if !unicode.IsGraphic(r) {
				binary = true
			}
		}
		if i > 0 {
			fmt.Fprintf(&b, " ")
		}
		if binary {
			fmt.Fprintf(&b, "%#v", token)
		} else {
			s := string(token)
			s = strings.ReplaceAll(s, "\r", "\\r")
			s = strings.ReplaceAll(s, "\n", "\\n")
			fmt.Fprintf(&b, "%s", s)
		}
	}
	return b.String()
}

// Send sends serial command
func (c *BP35C2Client) send(in []byte) error {
	c.sendSeq++
	log.Printf("Send[%d]:%s", c.sendSeq, stringWithBinary(in))
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

	log.Printf("Read[%d]:%s", c.readSeq, stringWithBinary(line))
	line = bytes.TrimSuffix(line, []byte{'\r', '\n'})
	return line, err
}

// Version is ..
func (c BP35C2Client) Version() (string, error) {
	err := c.send([]byte("SKVER\r\n"))
	if err != nil {
		return "", err
	}

	// Echoback
	_, err = c.recv()
	if err != nil {
		return "", err
	}

	//EVER X.Y.Z
	r, err := c.recv()
	if err != nil {
		return "", err
	}

	if !bytes.HasPrefix(r, []byte("EVER")) {
		return "", fmt.Errorf("unexpected response [%s]", r)
	}

	tokens := bytes.Split(r, []byte{' '})
	if len(tokens) < 2 {
		return "", fmt.Errorf("version string not found")
	}
	ver := string(tokens[1])

	//OK
	r, err = c.recv()
	if err != nil {
		return ver, err
	}
	if !bytes.Equal(r, []byte("OK")) {
		return ver, fmt.Errorf("command failed [%s]", r)
	}

	return ver, nil
}

// SetBRoutePassword is..
func (c BP35C2Client) SetBRoutePassword(password string) error {
	if len(password) == 0 {
		return fmt.Errorf("B-route password is empty")
	}

	c.send([]byte("SKSETPWD C " + password + "\r\n"))
	c.recv()
	c.recv()
	return nil
}

// SetBRouteID  is ..
func (c BP35C2Client) SetBRouteID(id string) error {
	if len(id) == 0 {
		return fmt.Errorf("B-route ID is empty")
	}

	c.send([]byte("SKSETRBID " + id + "\r\n"))
	c.recv()
	c.recv()
	return nil
}

func (c BP35C2Client) scan(duration int) bool {
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
				return len(data) != 0
			}
		case <-ctx.Done():
			log.Fatal(ctx.Err())
			return false
		}
	}
}

func (c BP35C2Client) receivePanDesc() (PanDesc, error) {
	ed := PanDesc{}
	line, err := c.recv()
	if err == nil && bytes.HasPrefix(line, []byte("EPANDESC")) {
		line, err := c.recv() // Channel
		if err != nil {
			return PanDesc{}, fmt.Errorf("failed to get Channel [%s]", err)
		}
		tokens := bytes.Split(line, []byte("Channel:"))
		ed.Channel = string(bytes.Trim(tokens[1], "\r\n"))
		c.recv()             // Channel Page: XX
		line, err = c.recv() // Pan ID: XXXX
		if err != nil {
			return PanDesc{}, fmt.Errorf("failed to get Pan ID [%s]", err)
		}
		tokens = bytes.Split(line, []byte("Pan ID:"))
		ed.PanID = string(bytes.Trim(tokens[1], "\r\n"))
		line, err = c.recv() // Addr:XXXXXXXXXXXXXXXX
		if err != nil {
			return PanDesc{}, fmt.Errorf("failed to get Addr [%s]", err)
		}
		tokens = bytes.Split(line, []byte("Addr:"))
		ed.Addr = string(bytes.Trim(tokens[1], "\r\n"))
		c.recv() // LQI:CA
		c.recv() // Side:X
		c.recv() // PairID:XXXXXXXX
	}
	return ed, err
}

// Scan is ..
func (c BP35C2Client) Scan() (PanDesc, error) {
	duration := 4
	for {
		if duration > 8 {
			log.Println("duration limit(8) exceeds")
			break
		}

		found := c.scan(duration)
		if found {
			break
		}
		duration = duration + 1
	}

	ed, err := c.receivePanDesc()
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
func (c *BP35C2Client) Join(desc PanDesc) (bool, error) {
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

			tokens := bytes.Split(res, []byte{' '})
			eventType := string(tokens[0])

			switch eventType {
			case "EVENT":
				if len(tokens) < 2 {
					return false, fmt.Errorf("invalid format [%s]", res)
				}
				num, err := strconv.ParseInt(string(tokens[1]), 16, 8)
				if err != nil {
					return false, fmt.Errorf("invalid EVENT num [%s]", res)
				}
				switch num {
				case 0x24:
					log.Println("Join failed")
					return false, nil
				case 0x25:
					log.Println("Join succeed")
					c.joined = true
					return true, nil
				}
			}
		}
	}

}

// Send is...
func (c *BP35C2Client) Send(data []byte) ([]byte, error) {
	ipv6 := c.panDesc.IPV6Addr
	cmd := []byte(fmt.Sprintf("SKSENDTO 1 %s 0E1A 1 0 %04X ", ipv6, len(data)))
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

			tokens := bytes.Split(res, []byte{' '})
			eventType := string(tokens[0])

			switch eventType {
			case "EVENT":
				if len(tokens) < 2 {
					log.Printf("invalid format [%s]\n", res)
				}
				num, err := strconv.ParseInt(string(tokens[1]), 16, 8)
				if err != nil {
					log.Printf("invalid EVENT num [%s]\n", res)
				}
				switch num {
				case 0x21:
					log.Println("UDP send succeed")
				default:
					log.Printf("unexpected EVENT %x\n", num)
				}
			case "ERXUDP":
				// ERXUDP <SENDER> <DEST> <RPORT> <LPORT> <SENDERLLA> (<RSSI>) <SECURED> <SIDE> <DATALEN> <DATA>
				if len(tokens) >= 10 {
					dstPort, err := strconv.ParseInt(string(tokens[4]), 16, 16)
					if err != nil {
						return nil, fmt.Errorf("invalid destination port [%s]", res)
					}
					switch dstPort {
					case 3610: // ECHONET Lite
						data := tokens[9]
						return data, err
					case 716: // PANA
						log.Println("PANA data")
					case 19788: // MLE
						log.Println("MLE data")
					}

				}
			}
		}
	}
}

// Connect connects to smart-meter
func (c *BP35C2Client) Connect(bRouteID, bRoutePW string) error {

	if len(bRouteID) == 0 {
		log.Fatal("set B-route ID")
	}
	if len(bRoutePW) == 0 {
		log.Fatal("set B-route password")
	}

	c.SetBRoutePassword(bRoutePW)
	c.SetBRouteID(bRouteID)

	pd, err := c.Scan()
	if err != nil {
		log.Fatal(err)
	}

	ipv6Addr, err := c.LL64(pd.Addr)
	if err != nil {
		log.Fatal(err)
	}

	pd.IPV6Addr = ipv6Addr
	log.Printf("Translated address:%#v", pd)

	err = c.SRegS2(pd.Channel)
	if err != nil {
		log.Fatal(err)
	}

	err = c.SRegS3(pd.PanID)
	if err != nil {
		log.Fatal(err)
	}

	// PANA authentication
	joined, err := c.Join(pd)
	if err != nil {
		log.Fatal(err)
	}

	if !joined {
		log.Fatal("Failed to join")
	}

	c.panDesc = pd

	// TODO: return error
	return nil
}

// Term terminates PANA session
func (c BP35C2Client) Term() {
	c.send([]byte("SKTERM\r\n"))
	c.recv()
	c.recv()
}
