package main

import (
	"fmt"
	"log"
)

func main() {
	var cfg = cmdlineConfig()
	if err := cfg.Parse(nil); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", optionsVar.OutputDirectory)
}
