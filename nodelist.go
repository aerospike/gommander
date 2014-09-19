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

func (l NodeList) Drop(n int) NodeList {
	return l[n:len(l)]
}

func (l NodeList) Take(n int) NodeList {
	if n > len(l) {
		n = len(l)
	}
	return l[0:n]
}

func (l NodeList) Filter(fn func(*Node) bool) NodeList {
	var p []*Node
	for _, v := range l {
		if fn(v) {
			p = append(p, v)
		}
	}
	return p
}

func (l NodeList) Connect() error {

	for _, n := range l {
		if err := n.Connect(); err != nil {
			return err
		}
	}

	return nil
}

func (l NodeList) Close() error {

	for _, n := range l {
		if err := n.Close(); err != nil {
			return err
		}
	}

	return nil
}

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

func (l NodeList) Execute(req Request) (chan Response, error) {
	return l.Each(func(n *Node, respond func(Response) error) error {
		req.Respond = respond
		return n.Execute(req)
	})
}

func (l NodeList) Run(command string) (chan Response, error) {
	req := Request{
		Command: command,
		Stdin:   new(bytes.Buffer),
	}

	return l.Execute(req)
}

func (l NodeList) Copy(src string, dest string) (chan Response, error) {

	command := "/usr/bin/scp -qrt ./"
	input := new(bytes.Buffer)

	in, err := os.Open(src)
	if err != nil {
		return nil, err
	}

	defer in.Close()

	inStat, err := in.Stat()
	if err != nil {
		return nil, err
	}

	fmt.Fprintln(input, "C0644", inStat.Size(), dest)
	if _, err = io.Copy(input, in); err != nil {
		return nil, err
	}
	fmt.Fprint(input, "\x00")

	return l.Each(func(n *Node, respond func(Response) error) error {

		req := Request{
			Command: command,
			Stdin:   bytes.NewReader(input.Bytes()),
			Respond: respond,
		}

		return n.Execute(req)
	})
}

func (l NodeList) Write(dest string, content *bytes.Reader) (chan Response, error) {

	command := "/usr/bin/scp -qrt ./"
	stdin := new(bytes.Buffer)

	fmt.Fprintln(stdin, "C0644", content.Len(), dest)
	if _, err := io.Copy(stdin, content); err != nil {
		return nil, err
	}
	fmt.Fprint(stdin, "\x00")

	return l.Each(func(n *Node, respond func(Response) error) error {

		req := Request{
			Command: command,
			Stdin:   bytes.NewReader(stdin.Bytes()),
			Respond: respond,
		}

		return n.Execute(req)
	})
}
