// rot: ROT cypher brute force
// This application decodes a ROT-n encoded
// string using all alphabet chars possibilities including
// white space
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
	"fmt"
	"os"
)

func main() {
	var string1 string
	var inString string

	string1 = "abcdefghijklmnopqrstuvwxyz "
	if os.Args[0] == "" {
		fmt.Println("Usage: rot <string_to_decode>")
		os.Exit(1)
	}
	inString = os.Args[1]
	byteInString := []byte(inString)
	lenInString := len(inString)

	byteStr := []byte(string1)
	lenString1 := len(string1)

	for k := 0; k <= lenString1-1; k++ {
		fmt.Printf("ROT%ds ", k)
		for i := 0; i <= lenInString-1; i++ {
			j := seq(bytes.Index(byteStr, byteInString[i:i+1])+k, lenString1)
			fmt.Printf("%s", byteStr[j:j+1])
		}
		fmt.Println("")
	}

}

func seq(a, offset int) int {
	if a < offset {
		return a
	} else {
		return a - offset
	}
}
