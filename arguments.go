package cmdline

import "strings"

// Arguments holds a slice strings containing arguments to parse and implements
// a few helpers to tokenize the slice.
type Arguments []string

// Kind returns the current argument kind.
func (self Arguments) Kind(config *Config) (kind ArgumentKind) {
	if self.Eof() {
		return NoArgument
	}
	kind = TextArgument
	// in case of "-" as short and "--" as long, long wins.
	if strings.HasPrefix(self.Raw(), config.ShortPrefix) {
		kind = ShortArgument
	}
	if strings.HasPrefix(self.Raw(), config.LongPrefix) {
		kind = LongArgument
	}
	return
}

// ArgumentKind defines the kind of argument being parsed.
type ArgumentKind int

const (
	// NoArgument indicates no argument. It's returned if Arguments are empty.
	NoArgument ArgumentKind = iota
	// LongArgument indicates a token with a long option prefix.
	LongArgument
	// ShortArgument indicates a token with a short option prefix.
	ShortArgument
	// TextArgument indicates a text token with no prefix.
	TextArgument
)

// Raw returns current token, unmodified.
func (self Arguments) Raw() string {
	if self.Eof() {
		return ""
	}
	return self[0]
}

// Text returns current token stripped of any prefixes, depending on kind.
func (self Arguments) Text(config *Config) string {
	switch k := self.Kind(config); k {
	case ShortArgument:
		return string(self.Raw()[len(config.ShortPrefix):])
	case LongArgument:
		return string(self.Raw()[len(config.LongPrefix):])
	case TextArgument:
		return self.Raw()
	}
	return ""
}

// Next advances to next token.
func (self *Arguments) Next() *Arguments {
	if self.Eof() {
		return self
	}
	*self = (*self)[1:]
	return self
}

// End advances past all tokens, to EOF.
func (self *Arguments) End() { self = &Arguments{} }

// Eof returns true if there are no more tokens in Arguments.
func (self Arguments) Eof() bool { return len(self) == 0 }
