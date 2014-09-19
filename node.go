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
	"io"
	"io/ioutil"
	"strconv"
)

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

type Request struct {

	// Command to execute
	Command string

	// Stdin Buffer
	Stdin io.Reader

	// Response Channel
	Respond func(Response) error
}

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

type Error struct {
	message string
}

func (e *Error) Error() string {
	return e.message
}

func NewNode(host string, port uint, user string, auth []ssh.AuthMethod) *Node {
	return &Node{
		Host: host,
		Port: port,
		User: user,
		Auth: auth,
	}
}

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

func (n *Node) Close() error {

	if n.requests == nil {
		return &Error{message: "Not listening."}
	}

	close(n.requests)
	n.requests = nil

	n.client.Close()
	return nil
}

func (n *Node) Execute(req Request) error {
	n.requests <- req
	return nil
}

func (n *Node) listen() error {

	if n.requests != nil {
		return &Error{message: "Already listening."}
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
