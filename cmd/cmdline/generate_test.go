package main

import (
	"testing"

	"github.com/vedranvuk/bast"
)

func TestGenerate(t *testing.T) {

	const tagName = "testTag"

	var config = &GenerateConfig{
		TagName:    tagName,
		Packages:   []string{"./..."},
		PackageName: "main",
		PointerVars: true,
		OutputFile: "../../_testproject/cmd/testcmd/commands.go",
		Bast: &bast.Config{
			Dir: "../../_testproject",
		},
	}
	if err := config.Generate(); err != nil {
		t.Fatal(err)
	}
	_ = config

}
