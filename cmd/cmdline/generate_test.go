package main

import (
	"os"
	"testing"

	"github.com/vedranvuk/bast"
)

func TestGenerate(t *testing.T) {
	var dir, err = os.Getwd()
	if err != nil {
		t.Fatal("cannot get working dir")
	}
	t.Log("Working dir:", dir)

	const tagName = "testTag"

	var config = &GenerateConfig{
		TagName:    tagName,
		Packages:   []string{"./..."},
		PackageName: "main",
		PointerVars: true,
		OutputFile: "../../_testproject/cmd/testcmd/commands.go",
		BastConfig: &bast.Config{
			Dir: "../../_testproject",
		},
	}
	if err := config.Generate(); err != nil {
		t.Fatal(err)
	}
	_ = config

}
