package main

import (
	"github.com/sebnyberg/sttrouter/cmd"
	"log"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
