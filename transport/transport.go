package transport

//go:generate mockgen -source transport.go -destination transport_mock.go -package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"
)

type ReceiveResult struct {
	Data    []byte
	Address string
	Err     error
}

type MulticastReceiver interface {
	Start(ctx context.Context) <-chan ReceiveResult
}

type MulticastSender interface {
	Send(data []byte)
}

type UnicastReceiver interface {
	Start(ctx context.Context) <-chan ReceiveResult
}

type UDPMulticastReceiver struct {
	IP   string
	Port string
}

func (r *UDPMulticastReceiver) Start(ctx context.Context) <-chan ReceiveResult {
	results := make(chan ReceiveResult, 5)
	log.Println("Start to listen multicast udp ", r.IP, r.Port)

	go func() {
		defer close(results)
		address, err := net.ResolveUDPAddr("udp", r.IP+r.Port)
		log.Println("resolved:", address)
		if err != nil {
			results <- ReceiveResult{Err: fmt.Errorf("Error: [%s]", err)}
		}
		conn, err := net.ListenMulticastUDP("udp", nil, address)
		if err != nil {
			results <- ReceiveResult{Err: fmt.Errorf("Error: [%s]", err)}
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

type UDPMulticastSender struct {
	conn net.Conn
}

func NewUDPMulticastSender(IP, Port string) (*UDPMulticastSender, error) {
	conn, err := net.Dial("udp", IP+Port)
	if err != nil {
		return nil, fmt.Errorf("Write conn error: [%s]", err)
	}
	return &UDPMulticastSender{conn: conn}, nil
}

func (ums *UDPMulticastSender) Close() {
	ums.conn.Close()
}

func (ums *UDPMulticastSender) Send(data []byte) {
	length, err := ums.conn.Write(data)
	if err != nil {
		log.Println("Write error: ", err)
	}
	log.Println("written:", length)
}

type UDPUnicastReceiver struct {
	IP   string
	Port string
}

func (r *UDPUnicastReceiver) Start(ctx context.Context) <-chan ReceiveResult {
	results := make(chan ReceiveResult, 5)
	log.Println("Start to listen unicast udp ", r.IP, r.Port)

	go func() {
		defer close(results)
		udpAddr, err := net.ResolveUDPAddr("udp", r.IP+r.Port)
		if err != nil {
			results <- ReceiveResult{Err: fmt.Errorf("Unicast Error: [%s]", err)}
		}
		conn, err := net.ListenUDP("udp", udpAddr)
		if err != nil {
			results <- ReceiveResult{Err: fmt.Errorf("Unicast Error: [%s]", err)}
		}
		defer conn.Close()
		buffer := make([]byte, 1500)

		for {
			conn.SetDeadline(time.Now().Add(1 * time.Second))
			length, remoteAddress, err := conn.ReadFrom(buffer)
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
