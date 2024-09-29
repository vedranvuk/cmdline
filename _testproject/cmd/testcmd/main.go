package main

import (
	"log"

	"github.com/vedranvuk/cmdline/_testproject/pkg/models"
)

var config = new(models.Config)

func main() {
	if err := cmdlineConfig().Parse(nil); err != nil {
		log.Fatal(err)
	}
}
