package main

import (
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	logger := log.New()
	massa := NewMassa(logger)

	for {
		massa.Process()
		time.Sleep(1 * time.Minute)
	}

}
