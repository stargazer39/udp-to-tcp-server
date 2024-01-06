package main

import (
	"crypto/tls"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

const UDP_BUFFER_SIZE = 8192
const MAX_PACKET_SIZE = 8194

func main() {
	udpAddrStr := flag.String("u", ":51820", "UDP to addr")
	tcpAddrStr := flag.String("l", ":8088", "Listen addr")
	tlsEnabled := flag.Bool("tls", false, "Enable TLS")
	tlsCert := flag.String("tls-cert", "random", "Set cert path. default \"random\" will generate a random cert)")

	flag.Parse()

	var err error

	tcpAddr, err := net.ResolveTCPAddr("tcp4", *tcpAddrStr)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var listener net.Listener
	var tlsCertificate tls.Certificate

	if *tlsEnabled {
		if *tlsCert == "random" {
			cert, key := GenRandomCert()

			tlsCertificate, err = tls.X509KeyPair(cert, key)

			if err != nil {
				log.Fatal("Cannot be loaded the certificate.", err.Error())
			}
		} else {
			log.Fatal("No cert. bye")
		}

		listener, err = tls.Listen("tcp", *tcpAddrStr, &tls.Config{Certificates: []tls.Certificate{tlsCertificate}})
	} else {
		listener, err = net.ListenTCP("tcp", tcpAddr)
	}

	if err != nil {
		log.Fatal("Can't listen on port specified.", err.Error())
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	log.Println("listening on ", tcpAddr.String())

	for {
		conn, err := listener.Accept()

		log.Println("got connection")
		if err != nil {
			fmt.Println(err)
		}
		go manageConn(conn, *udpAddrStr)
	}
}

func manageConn(tConn net.Conn, udpAddrStr string) {
	udpAddr, err := net.ResolveUDPAddr("udp", udpAddrStr)

	if err != nil {
		fmt.Println(err)
		return
	}

	conn, err2 := net.DialUDP("udp", nil, udpAddr)

	if err2 != nil {
		log.Println(err)
		return
	}

	go func() {
		var buf [MAX_PACKET_SIZE]byte

		for {

			err, n := recvbuffer(tConn, buf[:])

			if err != nil {
				break
			}
			// log.Println("got with length", length)
			if _, err := conn.Write(buf[2:n]); err != nil {
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

func recvbuffer(conn net.Conn, buff []byte) (error, int) {

	if _, err := io.ReadAtLeast(conn, buff, 2); err != nil {
		return err, 0
	}

	length := binary.LittleEndian.Uint16(buff[:2])

	if length > MAX_PACKET_SIZE-2 {
		return fmt.Errorf("too big"), 0
	}

	if _, err := io.ReadAtLeast(conn, buff[2:length+2], int(length)); err != nil {
		return err, 0
	}

	return nil, int(length) + 2
}
