package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	udpAddrStr := flag.String("u", ":51280", "UDP to addr")
	tcpAddrStr := flag.String("l", ":8088", "Listen addr")

	flag.Parse()

	tcpAddr, err := net.ResolveTCPAddr("tcp4", *tcpAddrStr)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Start listening for TCP connections on the given address
	listener, err := net.ListenTCP("tcp", tcpAddr)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for {
		// Accept new connections
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
		}
		// Handle new connections in a Goroutine for concurrency
		go manageConn(conn, *udpAddrStr)
	}
}

func manageConn(tConn net.Conn, udpAddrStr string) {
	// Resolve the string address to a UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", udpAddrStr)

	if err != nil {
		fmt.Println(err)
		return
	}

	// Dial to the address with UDP
	conn, err2 := net.DialUDP("udp", nil, udpAddr)

	if err2 != nil {
		log.Println(err)
		return
	}

	go func() {
		for {
			buf := make([]byte, 1026)

			_, err := tConn.Read(buf)

			if err != nil {
				log.Println(err)
				return
			}

			length := binary.LittleEndian.Uint16(buf[:2])

			if _, err := conn.Write(buf[2:length]); err != nil {
				log.Println(err)
				return
			}
		}
	}()

	go func() {
		for {
			buf := make([]byte, 1026)
			uBuff := make([]byte, 1024)

			i, _, err := conn.ReadFromUDP(uBuff)

			if err != nil {
				log.Println(err)
				return
			}

			binary.LittleEndian.PutUint16(buf, uint16(i))

			buf = append(buf, uBuff...)

			if _, err := tConn.Write(buf); err != nil {
				log.Println(err)
				return
			}
		}
	}()
}
