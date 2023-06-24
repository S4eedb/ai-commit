package main

import (
	"log"
)

func main() {
	err := run(getDiff())
	if err != nil {
		log.Fatal(err)
	}
}
