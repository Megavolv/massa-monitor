package main

import (
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	logger := log.New()
	massa := NewMassa(logger)
	err := massa.CheckExecutable()
	if err != nil {
		logger.Error(err)
		return
	}

	for {
		massa.Process()
		time.Sleep(2 * time.Minute)
	}

}
