package main

import (
	"fmt"
	"github.com/op/go-logging"
	"os"
)

var log = logging.MustGetLogger("")
var logFormat = logging.MustStringFormatter("%{time:2006-01-02 15:04:05.000} [%{level}] %{message}")

var backendConsole = logging.NewLogBackend(os.Stdout, "", 0)

func initLog() {
	logging.SetFormatter(logFormat)

	logging.SetBackend(backendConsole)
}

func initLogFile(logFile string, level logging.Level) {
	file, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)

	if err != nil {
		log.Error("Failed to initialize log file: " + fmt.Sprint(err))

		return
	}

	backendFile := logging.NewLogBackend(file, "", 0)

	backendFileLeveled := logging.AddModuleLevel(backendFile)
	backendFileLeveled.SetLevel(level, "")

	backendConsoleLeveled := logging.AddModuleLevel(backendConsole)
	backendConsoleLeveled.SetLevel(level, "")

	logging.SetBackend(backendConsoleLeveled, backendFileLeveled)

	log.Debug("Successfully initialized log file")
}
