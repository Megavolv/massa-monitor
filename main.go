package main

import (
	"time"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

var loglevel *string = flag.String("loglevel", "info", "Log level - one of: ")

func main() {
	flag.Parse()

	logger := log.New()
	switch *loglevel {
	case "trace":
		logger.SetLevel(log.TraceLevel)
	case "info":
		logger.SetLevel(log.InfoLevel)
	case "warn":
		logger.SetLevel(log.WarnLevel)
	case "err":
		logger.SetLevel(log.ErrorLevel)
	}

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
