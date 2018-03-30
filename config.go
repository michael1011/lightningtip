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

	defaultRESTHost    = "0.0.0.0:8081"
	defaultTlsCertFile = ""
	defaultTlsKeyFile  = ""

	defaultAccessDomain = ""

	defaultTipExpiry = 3600

	defaultLndGRPCHost  = "localhost:10009"
	defaultLndCertFile  = "tls.cert"
	defaultMacaroonFile = "admin.macaroon"
)

type config struct {
	ConfigFile string `long:"config" Description:"Location of the config file"`

	LogFile  string `long:"logfile" Description:"Location of the log file"`
	LogLevel string `long:"loglevel" Description:"Log level: debug, info, warning, error"`

	RESTHost    string `long:"resthost" Description:"Host for the REST interface of LightningTip"`
	TlsCertFile string `long:"tlscertfile" Description:"Certificate for using LightningTip via HTTPS"`
	TlsKeyFile  string `long:"tlskeyfile" Description:"Certificate for using LightningTip via HTTPS"`

	AccessDomain string `long:"accessdomain" Description:"The domain you are using LightningTip from"`

	TipExpiry int64 `long:"tipexpiry" Description:"Invoice expiry time in seconds"`

	LND *backends.LND `group:"LND" namespace:"lnd"`
}

var cfg config

var backend backends.Backend

func initConfig() {
	lndDir := getDefaultLndDir()

	cfg = config{
		ConfigFile: defaultConfigFile,

		LogFile:  defaultLogFile,
		LogLevel: defaultLogLevel,

		RESTHost:    defaultRESTHost,
		TlsCertFile: defaultTlsCertFile,
		TlsKeyFile:  defaultTlsKeyFile,

		AccessDomain: defaultAccessDomain,

		TipExpiry: defaultTipExpiry,

		LND: &backends.LND{
			GRPCHost:     defaultLndGRPCHost,
			CertFile:     path.Join(lndDir, defaultLndCertFile),
			MacaroonFile: path.Join(lndDir, defaultMacaroonFile),
		},
	}

	// Ignore unknown flags the first time parsing command line flags to prevent showing the unknown flag error twice
	flags.NewParser(&cfg, flags.IgnoreUnknown).Parse()

	errFile := flags.IniParse(cfg.ConfigFile, &cfg)

	// Parse flags again to override config file
	_, err := flags.Parse(&cfg)

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

	// TODO: add more backend options like for example c-lighting and eclair
	backend = cfg.LND
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