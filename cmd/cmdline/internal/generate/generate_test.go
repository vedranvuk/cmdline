// Copyright 2023-2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package generate

import (
	"os"
	"testing"
)

func TestGenerate(t *testing.T) {
	os.Remove("../../_testproject/cmd/testcmd/commands.go")
	var dir, err = os.Getwd()
	if err != nil {
		t.Fatal("cannot get working dir")
	}
	t.Log("Working dir:", dir)

	const tagName = "cmdline"

	var config = Default()
	config.Packages =   []string{"./..."}
	config.PackageName = "main"
	config.OutputFile = "../../../../_testproject/cmd/testcmd/commands.go"
	config.BastConfig.Dir = "../../../../_testproject"
	config.Print = true

	if err := Generate(config); err != nil {
		t.Fatal(err)
	}

}
