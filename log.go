package main

import (
	"github.com/op/go-logging"
	"os"
)

var log = logging.MustGetLogger("lightningTip")
var logFormat = logging.MustStringFormatter("%{time:2006-01-02 15:04:05.000} [%{level}] %{message}")

func initLog() {
	logging.SetFormatter(logFormat)
	logging.SetLevel(logging.DEBUG, "")

	backendConsole := logging.NewLogBackend(os.Stdout, "", 0)

	logging.SetBackend(backendConsole)

	file, err := os.OpenFile("lightningTip.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)

	if err == nil {
		defer file.Close()

		backendFile := logging.NewLogBackend(file, "", 0)

		logging.SetBackend(backendConsole, backendFile)

		log.Debug("Successfully initialized log file")

	} else {
		log.Critical("Failed to initialize log file: ", err)
	}

}
