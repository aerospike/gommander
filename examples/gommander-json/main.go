// Copyright 2013-2014 Aerospike, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	. "github.com/aerospike/gommander"
	"golang.org/x/crypto/ssh"

	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

type Host struct {
	Host     string `json:"host"`
	Port     uint   `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	SSHkey   string `json:"sshkey"`
}

func main() {

	var err error

	hosts_file := flag.String("hosts", "hosts.json", "A JSON file listing the hosts.")
	commands_file := flag.String("commands", "commands.json", "A JSON file listing the commands to execute.")
	flag.Parse()

	hosts_json, err := ioutil.ReadFile(*hosts_file)
	if err != nil {
		fmt.Printf("Failed to read '%s'\n", *hosts_file)
		fmt.Println(err)
		os.Exit(1)
	}

	hosts := []Host{}
	if err = json.Unmarshal(hosts_json, &hosts); err != nil {
		fmt.Printf("error: Failed to parse '%s'\n", *hosts_file)
		fmt.Println(err)
		os.Exit(1)
	}

	commands_json, err := ioutil.ReadFile(*commands_file)
	if err != nil {
		fmt.Printf("Failed to read '%s'\n", *commands_file)
		fmt.Println(err)
		os.Exit(1)
	}

	var commandsi interface{}
	// commands := []CommandChoice{}
	if err = json.Unmarshal(commands_json, &commandsi); err != nil {
		fmt.Printf("error: Failed to parse '%s'\n", *commands_file)
		fmt.Println(err)
		os.Exit(1)
	}

	commands := commandsi.([]interface{})

	// define the list of nodes
	nodes := NodeList{}

	// add nodes
	for _, h := range hosts {
		auth := []ssh.AuthMethod{}

		if len(h.Password) > 0 {
			auth = append(auth, ssh.Password(h.Password))
		}

		if len(h.SSHkey) > 0 {
			signer, err := Parsekey(h.SSHkey)
			if err != nil {
				panic(fmt.Sprintf("error: Failed to read SSH key '%s'", h.SSHkey))
			}
			auth = append(auth, ssh.PublicKeys(signer))
		}

		nodes = append(nodes, NewNode(h.Host, h.Port, h.Username, auth))
	}

	// connect to the nodes
	nodes.Connect()

	// setup executor
	executor := NewExecutor()

	// run the commands
	for _, c := range commands {

		cmd := c.(map[string]interface{})

		for k, v := range cmd {
			switch k {
			case "run":
				switch args := v.(type) {
				case string:
					desc := fmt.Sprintf("%s", args)
					executor.exec(desc, func() (chan Response, error) {
						return nodes.Run(v.(string))
					})
				default:
					panic("Command 'run' is invalid.")
				}
			case "copy":
				switch args := v.(type) {
				case map[string]interface{}:
					src := args["src"].(string)
					dest := args["dest"].(string)
					desc := fmt.Sprintf("copy %s => %s", src, dest)
					executor.exec(desc, func() (chan Response, error) {
						return nodes.Copy(src, dest)
					})
				default:
					panic("Command 'copy' is invalid.")
				}
			}
		}
	}

	// close connections to nodes
	nodes.Close()
}
