package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/michael1011/lightningtip/backends"
	"github.com/op/go-logging"
	"os/user"
	"path"
	"runtime"
	"strings"
)

const (
	defaultConfigFile = "lightningTip.conf"

	defaultLogFile  = "lightningTip.log"
	defaultLogLevel = "debug"

	defaultLndRPCHost   = "localhost:10009"
	defaultLndCertFile  = "tls.cert"
	defaultMacaroonFile = "admin.macaroon"
)

type config struct {
	ConfigFile string `long:"config" Description:"Config file location"`

	LogFile  string `long:"logfile" Description:"Log file location"`
	LogLevel string `long:"loglevel" Description:"Log level: debug, info, warning, error"`

	LND *backends.LND `group:"LND" namespace:"lnd"`
}

var cfg config

var backend backends.Backend
var backendName string

func initConfig() {
	lndDir := getDefaultLndDir()

	cfg = config{
		ConfigFile: defaultConfigFile,

		LogFile:  defaultLogFile,
		LogLevel: defaultLogLevel,

		LND: &backends.LND{
			RPCHost:      defaultLndRPCHost,
			CertFile:     path.Join(lndDir, defaultLndCertFile),
			MacaroonFile: path.Join(lndDir, defaultMacaroonFile),
		},
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
		log.Infof("Failed to parse config file: %v", errFile)
	}

	// TODO: add more backend options like for example c-lighting
	backend = cfg.LND
	backendName = "LND"
}

func getDefaultLndDir() (dir string) {
	usr, err := user.Current()

	if err == nil {
		switch runtime.GOOS {
		case "windows":
			dir = path.Join(usr.HomeDir, "AppData/Local/Lnd")

		case "darwin":
			dir = path.Join(usr.HomeDir, "Library/Application Support/Lnd/tls.cert")

		default:
			dir = path.Join(usr.HomeDir, ".lnd")
		}

	}

	return dir
}
