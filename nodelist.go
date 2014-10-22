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
	"fmt"
	"io"
	"os"
	"sync"
)

type NodeList []*Node

// Filter a NodeList based on a predicate function.
// The predicate function should return true, if the Node in the
// list should be kept. Otherwise false should be returned.
// The result is a new NodeList containing the Nodes which the
// predicate function returned true for.
func (l NodeList) Filter(fn func(*Node) bool) NodeList {
	var p []*Node
	for _, v := range l {
		if fn(v) {
			p = append(p, v)
		}
	}
	return p
}

// Connect each Node in NodeList to their respective servers.
func (l NodeList) Connect() error {

	for _, n := range l {
		if err := n.Connect(); err != nil {
			return err
		}
	}

	return nil
}

// Close the connections for each Node in the NodeList.
func (l NodeList) Close() error {

	for _, n := range l {
		if err := n.Close(); err != nil {
			return err
		}
	}

	return nil
}

// Perform an operation against each Node in the NodeList.
// The result will be channel of Responses for each Node in the NodeList.
func (l NodeList) Each(fn func(*Node, func(Response) error) error) (chan Response, error) {

	var wg sync.WaitGroup
	wg.Add(len(l))

	responses := make(chan Response)

	for _, n := range l {

		respond := func(res Response) error {
			responses <- res
			wg.Done()
			return nil
		}

		if err := fn(n, respond); err != nil {
			println(err)
		}
	}

	go func() {
		wg.Wait()
		close(responses)
	}()

	return responses, nil
}

// Execute a Request against each Node in the NodeList.
// The result will be channel of Responses for each Node in the NodeList.
func (l NodeList) Execute(req Request) (chan Response, error) {
	return l.Each(func(n *Node, respond func(Response) error) error {
		req.Respond = respond
		return n.Execute(req)
	})
}

// Run a command against each Node in the NodeList.
// The result will be channel of Responses for each Node in the NodeList.
func (l NodeList) Run(command string) (chan Response, error) {
	req := Request{
		Command: command,
		Stdin:   new(bytes.Buffer),
	}

	return l.Execute(req)
}

// Copy a file from src to dest on each Node.
// The result will be channel of Responses for each Node in the NodeList.
func (l NodeList) Copy(src string, dest string) (chan Response, error) {

	command := "cat - > " + dest
	input := new(bytes.Buffer)

	in, err := os.Open(src)
	if err != nil {
		return nil, err
	}

	defer in.Close()

	if _, err = io.Copy(input, in); err != nil {
		return nil, err
	}

	return l.Each(func(n *Node, respond func(Response) error) error {

		req := Request{
			Command: command,
			Stdin:   bytes.NewReader(input.Bytes()),
			Respond: respond,
		}

		return n.Execute(req)
	})
}

// Write a file from at dest on each Node.
// The result will be channel of Responses for each Node in the NodeList.
func (l NodeList) Write(dest string, content *bytes.Reader) (chan Response, error) {

	command := "cat - > " + dest
	stdin := new(bytes.Buffer)

	fmt.Fprintln(stdin, "C0755", content.Len(), dest)
	if _, err := io.Copy(stdin, content); err != nil {
		return nil, err
	}

	return l.Each(func(n *Node, respond func(Response) error) error {

		req := Request{
			Command: command,
			Stdin:   bytes.NewReader(stdin.Bytes()),
			Respond: respond,
		}

		return n.Execute(req)
	})
}
