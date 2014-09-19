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
	"bytes"
	"container/list"
	"fmt"
)

const (
	Color_Off string = "\033[0m"

	Black  string = "\033[0;30m"
	Red    string = "\033[0;31m"
	Green  string = "\033[0;32m"
	Yellow string = "\033[0;33m"
	Blue   string = "\033[0;34m"
	Purple string = "\033[0;35m"
	Cyan   string = "\033[0;36m"
	White  string = "\033[0;37m"

	BBlack  string = "\033[1;30m"
	BRed    string = "\033[1;31m"
	BGreen  string = "\033[1;32m"
	BYellow string = "\033[1;33m"
	BBlue   string = "\033[1;34m"
	BPurple string = "\033[1;35m"
	BCyan   string = "\033[1;36m"
	BWhite  string = "\033[1;37m"

	UBlack  string = "\033[4;30m"
	URed    string = "\033[4;31m"
	UGreen  string = "\033[4;32m"
	UYellow string = "\033[4;33m"
	UBlue   string = "\033[4;34m"
	UPurple string = "\033[4;35m"
	UCyan   string = "\033[4;36m"
	UWhite  string = "\033[4;37m"

	On_Black  string = "\033[40m"
	On_Red    string = "\033[41m"
	On_Green  string = "\033[42m"
	On_Yellow string = "\033[43m"
	On_Blue   string = "\033[44m"
	On_Purple string = "\033[45m"
	On_Cyan   string = "\033[46m"
	On_White  string = "\033[47m"

	IBlack  string = "\033[0;90m"
	IRed    string = "\033[0;91m"
	IGreen  string = "\033[0;92m"
	IYellow string = "\033[0;93m"
	IBlue   string = "\033[0;94m"
	IPurple string = "\033[0;95m"
	ICyan   string = "\033[0;96m"
	IWhite  string = "\033[0;97m"

	BIBlack  string = "\033[1;90m"
	BIRed    string = "\033[1;91m"
	BIGreen  string = "\033[1;92m"
	BIYellow string = "\033[1;93m"
	BIBlue   string = "\033[1;94m"
	BIPurple string = "\033[1;95m"
	BICyan   string = "\033[1;96m"
	BIWhite  string = "\033[1;97m"

	On_IBlack  string = "\033[0;100m"
	On_IRed    string = "\033[0;101m"
	On_IGreen  string = "\033[0;102m"
	On_IYellow string = "\033[0;103m"
	On_IBlue   string = "\033[0;104m"
	On_IPurple string = "\033[0;105m"
	On_ICyan   string = "\033[0;106m"
	On_IWhite  string = "\033[0;107m"
)

func color(s string, c string) string {
	return fmt.Sprintf("%s%s\033[0m", c, s)
}

func column(s string, w int) string {
	o := bytes.Buffer{}
	i := 0
	for ; i < len(s); i++ {
		o.WriteRune(rune(s[i]))
		if i >= w {
			break
		}
	}
	for j := i; j < w; j++ {
		o.WriteRune(' ')
	}
	return o.String()
}

type ColorMap struct {
	mapping map[string]string
	colors  *list.List
}

func NewColorMap() *ColorMap {
	colors := list.New()
	// colors.PushBack(Black)
	colors.PushBack(Red)
	colors.PushBack(Green)
	colors.PushBack(Yellow)
	colors.PushBack(Blue)
	colors.PushBack(Purple)
	colors.PushBack(Cyan)
	colors.PushBack(White)
	colors.PushBack(BBlack)
	colors.PushBack(BRed)
	colors.PushBack(BGreen)
	colors.PushBack(BYellow)
	colors.PushBack(BBlue)
	colors.PushBack(BPurple)
	colors.PushBack(BCyan)
	colors.PushBack(BWhite)
	colors.PushBack(UBlack)
	colors.PushBack(URed)
	colors.PushBack(UGreen)
	colors.PushBack(UYellow)
	colors.PushBack(UBlue)
	colors.PushBack(UPurple)
	colors.PushBack(UCyan)
	colors.PushBack(UWhite)
	// colors.PushBack(On_Black)
	// colors.PushBack(On_Red)
	// colors.PushBack(On_Green)
	// colors.PushBack(On_Yellow)
	// colors.PushBack(On_Blue)
	// colors.PushBack(On_Purple)
	// colors.PushBack(On_Cyan)
	// colors.PushBack(On_White)
	colors.PushBack(IBlack)
	colors.PushBack(IRed)
	colors.PushBack(IGreen)
	colors.PushBack(IYellow)
	colors.PushBack(IBlue)
	colors.PushBack(IPurple)
	colors.PushBack(ICyan)
	colors.PushBack(IWhite)
	colors.PushBack(BIBlack)
	colors.PushBack(BIRed)
	colors.PushBack(BIGreen)
	colors.PushBack(BIYellow)
	colors.PushBack(BIBlue)
	colors.PushBack(BIPurple)
	colors.PushBack(BICyan)
	colors.PushBack(BIWhite)
	// colors.PushBack(On_IBlack)
	// colors.PushBack(On_IRed)
	// colors.PushBack(On_IGreen)
	// colors.PushBack(On_IYellow)
	// colors.PushBack(On_IBlue)
	// colors.PushBack(On_IPurple)
	// colors.PushBack(On_ICyan)
	// colors.PushBack(On_IWhite)

	return &ColorMap{
		mapping: map[string]string{},
		colors:  colors,
	}
}

func (m *ColorMap) Get(key string) string {

	color, ok := m.mapping[key]
	if ok {
		return color
	}

	elem := m.colors.Front()
	if elem != nil {
		m.colors.Remove(elem)
		color = elem.Value.(string)
		m.mapping[key] = color
		return color
	}

	return Red
}
