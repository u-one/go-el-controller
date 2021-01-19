package transport

//go:generate mockgen -source transport.go -destination transport_mock.go -package transport

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"
)

// ReceiveResult is response data
type ReceiveResult struct {
	Data    []byte
	Address string
	Err     error
}

// MulticastReceiver is multicast receiver
type MulticastReceiver interface {
	Start(ctx context.Context, ip, port string) <-chan ReceiveResult
}

// MulticastSender is multicast sender
type MulticastSender interface {
	Send(data []byte)
	Close()
}

// UnicastReceiver is unicast receiver
type UnicastReceiver interface {
	Start(ctx context.Context, port string) <-chan ReceiveResult
}

// UDPMulticastReceiver is udp multicast receiver
type UDPMulticastReceiver struct {
}

// Start starts to receive
func (r *UDPMulticastReceiver) Start(ctx context.Context, ip, port string) <-chan ReceiveResult {
	results := make(chan ReceiveResult, 5)
	log.Println("Start to listen multicast udp ", ip, port)

	go func() {
		defer close(results)
		address, err := net.ResolveUDPAddr("udp", ip+port)
		log.Println("resolved:", address)
		if err != nil {
			results <- ReceiveResult{Err: fmt.Errorf("Error: [%s]", err)}
			return
		}
		conn, err := net.ListenMulticastUDP("udp", nil, address)
		if err != nil {
			results <- ReceiveResult{Err: fmt.Errorf("Error: [%s]", err)}
			return
		}
		defer conn.Close()
		buffer := make([]byte, 1500)

		for {
			fmt.Printf(".")
			conn.SetDeadline(time.Now().Add(1 * time.Second))
			length, remoteAddress, err := conn.ReadFromUDP(buffer)
			if err != nil {
				err, ok := err.(net.Error)
				if !ok || !err.Timeout() {
					results <- ReceiveResult{Err: fmt.Errorf("Error: [%s]", err)}
				}
			} else if length > 0 {
				fmt.Println()
				// Need copy because buffer will be cleared and reuse
				data := append([]byte{}, buffer[:length]...)
				results <- ReceiveResult{Data: data, Address: remoteAddress.IP.String(), Err: nil}
			}
			select {
			case <-ctx.Done():
				log.Println("ctx.Done")
				return
			default:
				//log.Println("recv: ", length)
			}

			for i := range buffer {
				buffer[i] = 0
			}
		}
	}()
	return results
}

// UDPMulticastSender is udp multicast sender
type UDPMulticastSender struct {
	conn net.Conn
}

// NewUDPMulticastSender creates DPMulticastSender instance
func NewUDPMulticastSender(IP, Port string) (*UDPMulticastSender, error) {
	conn, err := net.Dial("udp", IP+Port)
	if err != nil {
		return nil, fmt.Errorf("Write conn error: [%s]", err)
	}
	return &UDPMulticastSender{conn: conn}, nil
}

// Close closes connection
func (ums *UDPMulticastSender) Close() {
	ums.conn.Close()
}

// Send sends data
func (ums *UDPMulticastSender) Send(data []byte) {
	_, err := ums.conn.Write(data)
	if err != nil {
		log.Println("Write error: ", err)
	}
	//log.Println("written:", length)
}

// UDPUnicastReceiver is udp unicast receiver
type UDPUnicastReceiver struct {
}

// Start starts to receive
func (r *UDPUnicastReceiver) Start(ctx context.Context, port string) <-chan ReceiveResult {
	results := make(chan ReceiveResult, 5)
	log.Println("Start to listen unicast udp ", port)

	go func() {
		address, err := net.ResolveUDPAddr("udp", port)
		log.Println("resolved:", address)
		if err != nil {
			results <- ReceiveResult{Err: fmt.Errorf("Error: [%s]", err)}
			return
		}
		conn, err := net.ListenUDP("udp", address)
		if err != nil {
			results <- ReceiveResult{Err: fmt.Errorf("Unicast Error: [%s]", err)}
			return
		}
		defer conn.Close()

		err = resolveSocketOption(conn)
		if err != nil {
			results <- ReceiveResult{Err: fmt.Errorf("unicast error: %w", err)}
			return
		}

		buffer := make([]byte, 1024)
		for {
			length, remoteAddress, err := conn.ReadFromUDP(buffer)
			if err != nil {
				log.Println("Unicast Error:", err)
			} else if length > 0 {
				fmt.Println()
				// Need copy because buffer will be cleared and reuse
				data := append([]byte{}, buffer[:length]...)
				results <- ReceiveResult{Data: data, Address: remoteAddress.String(), Err: nil}
			}

			select {
			case <-ctx.Done():
				log.Println("ctx.Done")
				return
			default:
			}

			for i := range buffer {
				buffer[i] = 0
			}

		}
	}()
	return results
}
