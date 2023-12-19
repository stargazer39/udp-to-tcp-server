package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

const UDP_BUFFER_SIZE = 64 * 1024 * 1024

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

	log.Println("listening on ", tcpAddr.String())

	for {
		// Accept new connections
		conn, err := listener.Accept()

		log.Println("got connection")
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
			var buf [UDP_BUFFER_SIZE + 2]byte

			_, err := io.ReadFull(tConn, buf[:])

			// _, err := tConn.Read(buf)

			if err != nil {
				log.Println(err)
				continue
			}

			length := binary.LittleEndian.Uint16(buf[0:2])

			// log.Println("got with length", length)
			if _, err := conn.Write(buf[2 : length+2]); err != nil {
				log.Println(err)
				continue
			}

			// log.Println("wrote with length", length)
		}
	}()

	go func() {
		for {
			var buf []byte
			uBuff := make([]byte, UDP_BUFFER_SIZE)
			leg := make([]byte, 2)

			i, _, err := conn.ReadFromUDP(uBuff)
			// log.Println("read udp")

			if err != nil {
				log.Println(err)
				continue
			}

			binary.LittleEndian.PutUint16(leg, uint16(i))

			buf = append(buf, leg...)
			buf = append(buf, uBuff...)

			if _, err := tConn.Write(buf); err != nil {
				log.Println(err)
				continue
			}

			// log.Println("wrote udp buf")
		}
	}()
}
