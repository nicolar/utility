// smtp-proxy: tcp proxy and SMTP logger
// This app receives byte streams from client, understand L7 protocol SMTP
// logs the communication to disk and sends the byte stream 1:1 to a destination
// server
//
// Copyright 2020 Nicola Ruggero
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"runtime/debug"
	"time"

	uuid "github.com/hashicorp/go-uuid"
)

// Globals
const swVer = "1.0"

var verbose bool = false

// rectifier plugins structure
type rectifier struct {
	req, res []byte
	sendback bool
	desc     string
}

func logVerboseln(v ...interface{}) {
	if verbose {
		log.Println(v...)
	}
}

func logVerbosef(format string, v ...interface{}) {
	if verbose {
		log.Printf(format, v...)
	}
}

func main() {

	// Command line options
	localAddr := flag.String("local", ":25", "local address")
	remoteAddr := flag.String("remote", ":2525", "remote address")
	logFile := flag.String("log", "smtp-proxy.log", "log file")
	verboseFlag := flag.Bool("verbose", false, "Print additional information")
	showSwVer := flag.Bool("version", false, "Print software version and exit")
	flag.Parse()

	// Show Software version
	if *showSwVer {
		fmt.Printf("smtp-proxy: tcp proxy and SMTP logger\n")
		fmt.Printf("Version: %s\n", swVer)
		os.Exit(1)
	}

	// Assign globally
	verbose = *verboseFlag

	log.Printf("Starting ldap-proxy: tcp proxy and SMTP logger\n")
	log.Printf("Version: %s\n", swVer)
	log.Printf("Sending logs to: %s\n", *logFile)

	// Open log file for writing
	f, err := os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal("Unable to open log file: ", err)
	}
	defer f.Close()

	// Initialize logging
	log.SetOutput(f)
	// Rewrite software info to logfile
	log.Printf("Starting ldap-proxy: tcp proxy and SMTP logger\n")
	log.Printf("Version: %s\n", swVer)

	// Listen for connections
	ln, err := net.Listen("tcp", *localAddr)
	if err != nil {
		log.Fatal("Unable to create listener:", err)
	}
	defer ln.Close()
	log.Println("Listening from: ", *localAddr)
	log.Println("Sending to: ", *remoteAddr)

	// Accept new incoming connections
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		// Start a new thread to handle the new incoming connection
		go handleConn(conn, *remoteAddr)
	}
}

func generateRandomUUID(size int) (string, error) {
	connUUID, err := uuid.GenerateRandomBytes(size)
	return hex.EncodeToString(connUUID), err
}

// Handle new incoming connections, analyze L7 SMTP protocol, and proxy data
// to destination server
func handleConn(conn net.Conn, remoteAddr string) {

	connUUID, err := generateRandomUUID(4)
	if err != nil {
		log.Printf("Unable to generate UUID for a new incoming connection: %s - Error: %s\n", conn.RemoteAddr(), err.Error())
		conn.Close()
		return
	}

	log.Printf("[%s] New connection from: %s\n", connUUID, conn.RemoteAddr())

	// Connect to remote server to proxy data to
	rconn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		log.Printf("[%s] Error dialing %s", connUUID, err.Error())
		conn.Close()
		return
	}
	log.Printf("[%s] Established server connection to: %s\n", connUUID, rconn.RemoteAddr())

	// Start 2 new threads to handle the requests/responses inside the connection
	// we need 2 async threads otherwise an incomplete request/response
	// may block the communication flow from the OSI L7 perspective
	// because of infinite waiting for data from one of the counterparts
	go handleRequest(conn, rconn, "client to proxy", connUUID) // client to proxy
	go handleRequest(rconn, conn, "server to proxy", connUUID) // server to proxy
}

func handleRequest(conn net.Conn, rconn net.Conn, desc string, connUUID string) {
	defer func() {
		if r := recover(); r != nil {
			logVerbosef("[%s] Recovering from panic: %s\n", connUUID, r)
			logVerbosef("[%s] Stack Trace:\n", connUUID)
			if verbose {
				debug.PrintStack()
			}
		}
		logVerbosef("[%s] handleRequest: deferred connection closure: %s\n", connUUID, desc)
		conn.Close()
		rconn.Close()
	}()

	// Initialize read buffer
	// RFC 5321 4.5.3.1.4 - max size for command is 512 octets, let's double it
	buf := make([]byte, 1024)

	// Loop while communication channel is alive
	for {
		// Read SMTP data from source
		start := time.Now()
		logVerbosef("[%s]   conn.Read -> %s\n", connUUID, desc)
		packetLen, err := conn.Read(buf)
		if err != nil {
			// Don't log EOF errors
			if err != io.EOF {
				log.Printf("[%s]   Error read: %s\n", connUUID, err.Error())
			}
			return
		}
		t := time.Now()
		elapsed := t.Sub(start)
		logVerbosef("[%s]   Duration conn.Read -> %s %s", connUUID, desc, elapsed)
		b := buf[:packetLen]

		// Dump buffer to log for debug
		logVerbosef("[%s]   Received %d bytes: %s\n", connUUID, packetLen, desc)
		logVerbosef("[%s]   HEXDUMP:\n%s", connUUID, hex.Dump(b[:packetLen]))

		// Extract sender and recipients
		if bytes.HasPrefix(b, []byte("MAIL FROM:")) {
			log.Printf("[%s]   MAIL FROM: %s\n", connUUID, string(bytes.TrimPrefix(b, []byte("MAIL FROM:"))))
		}
		if bytes.HasPrefix(b, []byte("RCPT TO:")) {
			log.Printf("[%s]   RCPT TO: %s\n", connUUID, string(bytes.TrimPrefix(b, []byte("RCPT TO:"))))
		}

		// Write SMTP data to destination
		start = time.Now()
		logVerbosef("[%s]   conn.Write -> %s\n", connUUID, desc)
		_, err = rconn.Write(b)
		if err != nil {
			log.Printf("[%s]   Error write: %s\n", connUUID, err.Error())
			return
		}
		t = time.Now()
		elapsed = t.Sub(start)
		logVerbosef("[%s]   Duration conn.Write -> %s %s", connUUID, desc, elapsed)
	}
}
