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

package gommander

import (
	"bytes"
	"code.google.com/p/go.crypto/ssh"
	"errors"
	"io"
	"io/ioutil"
	"strconv"
)

// Node describes a server which will be managed.
type Node struct {

	// Address of the node
	Host string

	// Port of the node
	Port uint

	// Username for user to authenticate
	User string

	// Authentication Method
	Auth []ssh.AuthMethod

	// SSH Client
	client *ssh.Client

	requests chan Request
}

// Request represents the command, stdin and callback responder
// to be executed by a Node.
type Request struct {

	// Command to execute
	Command string

	// Stdin Buffer
	Stdin io.Reader

	// Response Channel
	Respond func(Response) error
}

// Response represents the result of a command executed on a Node, including exit code,
// stdout and stderr.
type Response struct {

	// Node
	Node *Node

	// Exit Code
	ExitCode int

	// Stdout Buffer
	Stdout *bytes.Buffer

	// Stderr Buffer
	Stderr *bytes.Buffer
}

// Create a new Node.
// The []AuthMethod used for each Node, can be unique to the specific Node or
// share amongst Nodes.
func NewNode(host string, port uint, user string, auth []ssh.AuthMethod) *Node {
	return &Node{
		Host: host,
		Port: port,
		User: user,
		Auth: auth,
	}
}

// A utility function to simplify the reasing and parsing of SSH Private Keys.
func Parsekey(file string) (private ssh.Signer, err error) {

	privateBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	private, err = ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return
	}

	return
}

// Connect to the node over SSH.
func (n *Node) Connect() error {

	config := &ssh.ClientConfig{
		User: n.User,
		Auth: n.Auth,
	}

	client, err := ssh.Dial("tcp", n.Host+":"+strconv.Itoa(int(n.Port)), config)

	if err != nil {
		return err
	}

	n.client = client
	n.listen()

	return nil
}

// Close the SSH connection.
func (n *Node) Close() error {

	if n.requests == nil {
		return errors.New("Not listening.")
	}

	close(n.requests)
	n.requests = nil

	n.client.Close()
	return nil
}

// Execute a request against the Node.
// This sends the request to the Node receive channel, for processing.
func (n *Node) Execute(req Request) error {
	n.requests <- req
	return nil
}

// Listens for Requests on the Node receive channel, then processes
// the request and sends the Response to channel specific by the Request.
func (n *Node) listen() error {

	if n.requests != nil {
		return errors.New("Already listening.")
	}

	n.requests = make(chan Request)

	go func(n *Node) {
		for req := range n.requests {
			res := Response{
				Node:     n,
				ExitCode: 0,
				Stdout:   new(bytes.Buffer),
				Stderr:   new(bytes.Buffer),
			}
			n.execute(&req, &res)
			req.Respond(res)
		}
	}(n)
	return nil
}

// Execute a Request and populate the Response.
func (n *Node) execute(req *Request, res *Response) error {

	session, err := n.client.NewSession()
	if err != nil {
		return err
	}

	defer session.Close()

	session.Stdout = res.Stdout
	session.Stderr = res.Stderr

	go func() {
		stdin, err := session.StdinPipe()
		if err != nil {
			return
		}

		defer stdin.Close()

		io.Copy(stdin, req.Stdin)
	}()

	if err = session.Run(req.Command); err != nil {
		res.ExitCode = err.(*ssh.ExitError).ExitStatus()
	} else {
		res.ExitCode = 0
	}

	return nil
}
