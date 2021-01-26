// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmdline

import (
	"errors"
	"testing"
)

func getTestCommands() []*Command {
	return []*Command{
		NewCommand(nil, "one", "", nil),
		NewCommand(nil, "two", "", nil),
		NewCommand(nil, "three", "", nil),
	}
}

func TestBaseChain(t *testing.T) {
	var commands = getTestCommands()
	var chain = &BaseChain{commands}
	if chain.First() != commands[0] {
		t.Fatal("BaseChain.First failed.")
	}
	if chain.Last() != commands[2] {
		t.Fatal("BaseChain.Last failed.")
	}
	if chain.Index(1) != commands[1] {
		t.Fatal("BaseChain.Index failed.")
	}
	var i int
	chain.Visit(func(c *Command) bool {
		if c != commands[i] {
			t.Fatal("BaseChain.Visit failed.")
		}
		i++
		return true
	}, false)
	i = 2
	chain.Visit(func(c *Command) bool {
		if c != commands[i] {
			t.Fatal("BaseChain.Visit failed.")
		}
		i--
		return true
	}, true)
	i = 0
	chain.Visit(func(c *Command) bool {
		if i == 1 {
			return false
		}
		i++
		return true
	}, false)
	if i != 1 {
		t.Fatal("BaseChain.Visit failed.")
	}
	i = 2
	chain.Visit(func(c *Command) bool {
		if i == 1 {
			return false
		}
		i--
		return true
	}, true)
	if i != 1 {
		t.Fatal("BaseChain.Visit failed.")
	}
	chain = &BaseChain{}
	if chain.First() != nil {
		t.Fatal("BaseChain.First failed.")
	}
	if chain.Last() != nil {
		t.Fatal("BaseChain.Last failed.")
	}
	if chain.Index(1) != nil {
		t.Fatal("BaseChain.Index failed.")
	}
}

func TestStaticChain(t *testing.T) {
	var commands = getTestCommands()
	var chain = NewStaticChain(commands...)
	for i := 0; i < len(commands); i++ {
		if chain.Index(i) != commands[i] {
			t.Fatal("NewStaticChain failed.")
		}
	}
	chain = NewStaticChain()
	for i := 0; i < len(commands); i++ {
		chain.Add(commands[i])
	}
	for i := 0; i < len(commands); i++ {
		if chain.Index(i) != commands[i] {
			t.Fatal("StaticChain.Add failed.")
		}
	}
}

func TestActiveChain(t *testing.T) {
	var ok bool
	var propagated = errors.New("propagated")
	var command = NewCommand(nil, "test", "", func(ctx Context) error {
		ok = true
		return propagated
	})
	var chain = NewActiveChain()
	var err error
	if err = chain.Add(command); err != propagated {
		t.Fatal("result propagation failed")
	}
	if !ok {
		t.Fatal("handler did not fire")
	}
}
