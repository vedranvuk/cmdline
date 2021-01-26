// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmdline

import (
	"reflect"
	"strings"
)

// PrintCommands returns a print string of specified commands.
func PrintCommands(commands *Commands) string {
	var sb = strings.Builder{}
	printCommands(&sb, commands, 0)
	return sb.String()
}

// printCommands is a recursive printer or registered Commands and Parameters.
// Lines are written to sb from current commands with the indent depth(*tab).
func printCommands(sb *strings.Builder, commands *Commands, indent int) {
	for _, commandname := range commands.nameindexes {
		command := commands.commandmap[commandname]
		writeIndent(sb, indent)
		sb.WriteString(commandname)
		if command.help != "" {
			sb.WriteRune('\t')
			sb.WriteString(command.help)
		}
		sb.WriteRune('\n')
		for _, paramlong := range command.Parameters().longindexes {
			param := command.Parameters().longparams[paramlong]
			shortparam := command.Parameters().longtoshort[paramlong]
			writeIndent(sb, indent)
			sb.WriteRune('\t')
			if param.required {
				if !param.indexed {
					sb.WriteString("<--")
				} else {
					sb.WriteRune('<')
				}
				sb.WriteString(paramlong)
				sb.WriteRune('>')
			} else {
				if !param.indexed {
					sb.WriteString("[--")
				} else {
					sb.WriteRune('[')
				}
				sb.WriteString(paramlong)
				sb.WriteRune(']')
			}
			if shortparam != "" {
				sb.WriteString("\t-")
				sb.WriteString(shortparam)
			}
			if param.value != nil {
				sb.WriteString("\t(")
				sb.WriteString(reflect.Indirect(reflect.ValueOf(param.value)).Type().Kind().String())
				sb.WriteRune(')')
			}
			if param.help != "" {
				sb.WriteRune('\t')
				sb.WriteString(param.help)
			}
			sb.WriteRune('\n')
		}
		sb.WriteRune('\n')
		if len(command.Commands().commandmap) > 0 {
			printCommands(sb, command.Commands(), indent+1)
		}
	}
}

// writeIndent writes an indent string of n depth to sb.
func writeIndent(sb *strings.Builder, n int) {
	for i := 0; i < n; i++ {
		sb.WriteRune('\t')
	}
}
