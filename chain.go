// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmdline

// Chain manages a chain of commands.
type Chain interface {
	// Add adds a Command to Chain. 
	// Returned error depends on Chain implementation.
	Add(*Command) error
	// Length returns number of commands in the Chain.
	Length() int
	// First returns first Command in Chain or nil if Chain is empty.
	First() *Command
	// Last returns last Command in Chain or nil if Chain is empty.
	Last() *Command
	// Index returns nth Command in Chain or nil if n exceeds Length.
	Index(int) *Command
	// Visit visits all Commands in chain, optionally in reverse. 
	// Func must return true to continue iteration.
	Visit(func(*Command) bool, bool)
}

// StaticChain is a chain which does nothing when a Command is added.
type StaticChain struct{ BaseChain }

// NewStaticChain returns a new StaticChain optionally containing commands.
func NewStaticChain(commands ...*Command) *StaticChain {
	return &StaticChain{BaseChain{commands}}
}

// Add adds command to end of chain. Result is always nil.
func (s *StaticChain) Add(command *Command) error {
	s.chain = append(s.chain, command)
	return nil
}

// ActiveChain is like StaticChain but executes Command's 
// Handler after adding it to chain.
type ActiveChain struct{ BaseChain }

// NewActiveChain returns a new ActiveChain optionally containing commands.
func NewActiveChain(commands ...*Command) *ActiveChain {
	return &ActiveChain{BaseChain{commands}}
}

// Add adds command to end of chain, calls its Handler and returns its result.
// If Command has no handler result is nil.
func (ac *ActiveChain) Add(command *Command) error {
	var context = &context{
		Command: command,
	}
	ac.chain = append(ac.chain, command)
	if command.Handler() != nil {
		return command.Handler()(context)
	}
	return nil
}

// BaseChain is a base shared type implementing common Chain functionality.
type BaseChain struct{ chain []*Command }

// Length returns number of commands in chain.
func (m BaseChain) Length() int { return len(m.chain) }

// First returns first Command in chain.
func (s BaseChain) First() *Command {
	if s.Length() < 1 {
		return nil
	}
	return s.chain[0]
}

// Last returns last Command in chain.
func (s BaseChain) Last() *Command {
	if s.Length() < 1 {
		return nil
	}
	return s.chain[s.Length()-1]
}

// Index returns Command at index i which must be in range or result is nil.
func (s BaseChain) Index(i int) *Command {
	if i >= s.Length() {
		return nil
	}
	return s.chain[i]
}

// Visit visits all Commands in chain by calling f for each command in chain.
// If f returns false iteration is aborted. If reverse, starts at last Command.
func (s BaseChain) Visit(f func(c *Command) bool, reverse bool) {
	if reverse {
		for i := s.Length() - 1; i >= 0; i-- {
			if !f(s.chain[i]) {
				break
			}
		}
		return
	}
	for i := 0; i < s.Length(); i++ {
		if !f(s.chain[i]) {
			break
		}
	}
}
