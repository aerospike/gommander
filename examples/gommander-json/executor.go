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

	"fmt"
	"strings"
	"time"
)

type Executor struct {
	colorMap *ColorMap
}

func NewExecutor() *Executor {
	return &Executor{
		colorMap: NewColorMap(),
	}
}

func (e *Executor) printResponses(responses chan Response, err error) {

	const layout = "2006-01-02 15:04:05 -0700"
	outtag := color(column(" OUT", 5), IBlack)
	errtag := color(column(" ERR", 5), Yellow+BIRed)
	septag := color("]", IBlack)

	if err != nil {
		panic(err)
	}

	for r := range responses {

		println(color("--------------------------------------------------------------------------------", IBlack))

		exittag := color(column(fmt.Sprintf("%d", r.ExitCode), 3), White)
		datetime := time.Now()
		datetimetag := color(datetime.Format(layout), IBlack)
		hosttag := color(column(r.Node.Host, 15), e.colorMap.Get(r.Node.Host))
		epre := fmt.Sprintf("%s  %20s %s %s %s  ", datetimetag, hosttag, errtag, exittag, septag)
		opre := fmt.Sprintf("%s  %20s %s %s %s  ", datetimetag, hosttag, outtag, exittag, septag)

		ostr := strings.Trim(r.Stdout.String(), "\r\n ")
		if len(ostr) > 0 {
			for _, s := range strings.Split(ostr, "\n") {
				fmt.Println(opre + s)
			}
		} else {
			if r.ExitCode == 0 {
				fmt.Println(opre + "<SUCCESS>")
			}
		}

		estr := strings.Trim(r.Stderr.String(), "\r\n ")
		if len(estr) > 0 {
			for _, s := range strings.Split(estr, "\n") {
				fmt.Println(epre + s)
			}
		} else {
			if r.ExitCode != 0 {
				fmt.Println(epre + "<FAILURE>")
			}
		}
	}
}

func (e *Executor) exec(desc string, fn func() (chan Response, error)) {

	println("################################################################################")
	println(">", desc)
	resp, err := fn()
	e.printResponses(resp, err)
}
