package main

import (
	log "log"

	wired "wired/src"
)

func main() {
	if err := wired.Run(); err != nil {
		log.Fatal(err)
	}
}
