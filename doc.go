// Copyright 2023-2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

/*
Package cmdline implements a command line parser.

It supports structured command invocation, several option types and mapping to 
basic variable types.

Design revolves around athe central [Config] type that defines everything from
arguments to parse, commands to handle them and their options and customizes 
the parse behaviour.

It mostly mimics GNU style option parsing with the addition of commands.

See [Config] and [Config.Parse] docs for more details.

*/
package cmdline
