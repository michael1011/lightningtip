package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/op/go-logging"
	"strings"
)

const (
	defaultConfigFile = "lightningTip.conf"

	defaultLogFile  = "lightningTip.log"
	defaultLogLevel = "debug"
)

type config struct {
	ConfigFile string `long:"config"  Description:"Config file location"`

	LogFile  string `long:"logfile" Description:"Log file location"`
	LogLevel string `long:"loglevel" Description:"Log level: debug, info, warning, error"`
}

var cfg config

func initConfig() {
	cfg = config{
		ConfigFile: defaultConfigFile,

		LogFile:  defaultLogFile,
		LogLevel: defaultLogLevel,
	}

	_, err := flags.Parse(&cfg)

	errFile := flags.IniParse(cfg.ConfigFile, &cfg)

	// Parse flags again to override config file
	_, err = flags.Parse(&cfg)

	// Default log level
	logLevel := logging.DEBUG

	switch strings.ToLower(cfg.LogLevel) {
	case "info":
		logLevel = logging.INFO

	case "warning":
		logLevel = logging.WARNING

	case "error":
		logLevel = logging.ERROR
	}

	initLogFile(cfg.LogFile, logLevel)

	if err != nil {
		log.Error("Failed to parse command line flags")
	}

	if errFile != nil {
		log.Infof("Could not parse config file: %v", errFile)
	}

}
