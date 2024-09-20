package main

import "log"

func main() {
	if err := cmdlineConfig().Parse(nil); err != nil {
		log.Fatal(err)
	}
}
