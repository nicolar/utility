// healthcheck: REST API to perform local healthchecks
// This is a REST API server that perform local healthchecks
// and provides a JSON answer
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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func doHealthCheck(w http.ResponseWriter, r *http.Request) {

	type subHealthCheck struct {
		HealthCheck       string `json:"healthCheck"`
		StatusCode        uint64 `json:"statusCode"`
		StatusDescription string `json:"statusDescription"`
	}

	type healthCheck struct {
		StatusCode        uint64           `json:"statusCode"`
		StatusDescription string           `json:"statusDescription"`
		HealthChecks      []subHealthCheck `json:"healthChecks"`
	}

	outb, err := exec.Command("./runme.sh", "1").Output()
	if err != nil {
		log.Println(err)
	}

	var sc uint64
	var sd string
	out := strings.TrimRight(string(outb), "\n")

	if value, _ := strconv.Atoi(out); value > 500 {
		sc = 1
		sd = "Too many documents to replicate: " + out
	} else {
		sc = 0
		sd = "OK: " + out
	}

	hc := healthCheck{
		StatusCode:        0,
		StatusDescription: "OK",
		HealthChecks: []subHealthCheck{
			subHealthCheck{
				HealthCheck:       "docsNotReplicated",
				StatusCode:        sc,
				StatusDescription: sd,
			},
		},
	}

	//fmt.Println(hc)
	json.NewEncoder(w).Encode(hc)
}

func main() {

	certPem := []byte(`-----BEGIN CERTIFICATE-----
-----END CERTIFICATE-----`)
	keyPem := []byte(`-----BEGIN RSA PRIVATE KEY-----
-----END RSA PRIVATE KEY-----`)
	cert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		log.Println("X509KeyPair")
		fmt.Printf("%s\n", certPem)
		fmt.Printf("%s\n", keyPem)
		log.Fatal(err)
	}
	cfg := &tls.Config{Certificates: []tls.Certificate{cert}}
	srv := &http.Server{
		Addr:         "0.0.0.0:9443",
		TLSConfig:    cfg,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Starting healthcheck")

	http.HandleFunc("/healthcheck", doHealthCheck)
	err = srv.ListenAndServeTLS("", "")
	if err != nil {
		log.Fatal("Unable to start healthcheck: ", err)
	}

}
