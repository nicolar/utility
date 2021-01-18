/*
tcp-server-echo: TCP echo server that echoes
data received by clients

Nicola Ruggero 2020 <nicola@nxnt.org>
*/
package main

import (
	"flag"
	"log"
	"net"
)

func main() {

	// Command line options
	localAddr := flag.String("local", ":4000", "local address")
	flag.Parse()

	log.Println("TCP Server ECHO")
	log.Println("Nicola Ruggero 2020 <nicola@nxnt.org>")

	// Listen for connections
	ln, err := net.Listen("tcp", *localAddr)
	if err != nil {
		log.Fatal("Unable to create listener:", err)
	}
	defer ln.Close()
	log.Println("Listening from: ", *localAddr)

	// Accept new incoming connections
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		// Start a new thread to handle the new incoming connection
		go handleConn(conn)
	}
}

// Handle new incoming connections sending back data received from server
func handleConn(conn net.Conn) {

	log.Println("New connection from: ", conn.RemoteAddr())

	defer func() {
		log.Println("Go routine: deferred connection closure")
		conn.Close()
	}()

	buff := make([]byte, 1024)

	for {
		// Read data from client
		n, err := conn.Read(buff)
		if err != nil {
			log.Println("Error read from client:", err)
			return
		}
		log.Printf("Received %d bytes from client\n", n)
		b := buff[:n]

		// Send data back to client
		_, err = conn.Write(b)
		if err != nil {
			log.Println("Error write to client:", err)
			return
		}
	}
}
