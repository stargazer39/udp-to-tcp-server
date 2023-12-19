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

const UDP_BUFFER_SIZE = 8192

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
			ln := make([]byte, 2)

			_, err := io.ReadFull(tConn, ln)

			// _, err := tConn.Read(buf)

			if err != nil {
				log.Println(err)
				continue
			}

			length := binary.LittleEndian.Uint16(ln)

			messageln := make([]byte, length)

			if _, err := io.ReadFull(tConn, messageln); err != nil {
				log.Println(err)
				continue
			}

			// log.Println("got with length", length)
			if _, err := conn.Write(messageln); err != nil {
				log.Println(err)
				continue
			}

			// log.Println("wrote with length", length)
		}
	}()

	go func() {
		for {
			// var buf []byte
			var uBuff []byte
			leg := make([]byte, 2)

			i, _, err := conn.ReadFromUDP(uBuff)
			// log.Println("read udp")

			if err != nil {
				log.Println(err)
				continue
			}

			binary.LittleEndian.PutUint16(leg, uint16(i))

			if _, err := tConn.Write(leg); err != nil {
				log.Println(err)
				continue
			}

			if _, err := tConn.Write(uBuff); err != nil {
				log.Println(err)
				continue
			}

			// log.Println("wrote udp buf")
		}
	}()
}
