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
	udpAddrStr := flag.String("u", ":51820", "UDP to addr")
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

			buf, err := recvbuffer(tConn)

			if err != nil {
				break
			}
			// log.Println("got with length", length)
			if _, err := conn.Write(buf); err != nil {
				log.Println(err)
				continue
			}

			// log.Println("wrote with length", length)
		}
	}()

	go func() {
		for {
			// var buf []byte
			var uBuff = make([]byte, UDP_BUFFER_SIZE)

			i, _, err := conn.ReadFromUDP(uBuff)
			// log.Println("read udp")

			if err != nil {
				log.Println(err)
				break
			}

			if err := sendBuffer(uBuff[:i], tConn); err != nil {
				log.Println(err)
				break
			}
			// log.Println("wrote udp buf")
		}
	}()
}

func sendBuffer(buffer []byte, conn net.Conn) error {
	length := make([]byte, 2)

	binary.LittleEndian.PutUint16(length, uint16(len(buffer)))

	i, err := conn.Write(length)

	if err != nil {
		return err
	}

	if i != len(length) {
		log.Fatal("len")
	}

	j, err := conn.Write(buffer)

	if err != nil {
		return err
	}

	if j != len(buffer) {
		log.Fatal("buf")
	}

	return nil
}

func recvbuffer(conn net.Conn) ([]byte, error) {
	length := make([]byte, 2)

	if _, err := io.ReadFull(conn, length); err != nil {
		return nil, err
	}

	msg := make([]byte, binary.LittleEndian.Uint16(length))

	if _, err := io.ReadFull(conn, msg); err != nil {
		return nil, err
	}

	return msg, nil
}
