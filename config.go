package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/michael1011/lightningtip/backends"
	"github.com/michael1011/lightningtip/database"
	"github.com/michael1011/lightningtip/notifications"
	"github.com/op/go-logging"
)

const (
	defaultConfigFile = "lightningTip.conf"

	defaultDataDir = "LightningTip"

	defaultLogFile  = "lightningTip.log"
	defaultLogLevel = "info"

	defaultDatabaseFile = "tips.db"

	defaultRESTHost    = "0.0.0.0:8081"
	defaultTLSCertFile = ""
	defaultTLSKeyFile  = ""

	defaultAccessDomain = ""

	defaultTipExpiry = 3600

	defaultReconnectInterval = 0
	defaultKeepaliveInterval = 0

	defaultLndGRPCHost  = "localhost:10009"
	defaultLndCertFile  = "tls.cert"
	defaultMacaroonFile = "invoice.macaroon"

	defaultRecipient = ""
	defaultSender    = ""

	defaultSTMPServer   = ""
	defaultSTMPSSL      = false
	defaultSTMPUser     = ""
	defaultSTMPPassword = ""
)

type config struct {
	ConfigFile string `long:"config" Description:"Location of the config file"`

	DataDir string `long:"datadir" Description:"Location of the data stored by LightningTip"`

	LogFile  string `long:"logfile" Description:"Location of the log file"`
	LogLevel string `long:"loglevel" Description:"Log level: debug, info, warning, error"`

	DatabaseFile string `long:"databasefile" Description:"Location of the database file to store settled invoices"`

	RESTHost    string `long:"resthost" Description:"Host for the REST interface of LightningTip"`
	TLSCertFile string `long:"tlscertfile" Description:"Certificate for using LightningTip via HTTPS"`
	TLSKeyFile  string `long:"tlskeyfile" Description:"Certificate for using LightningTip via HTTPS"`

	AccessDomain string `long:"accessdomain" Description:"The domain you are using LightningTip from"`

	TipExpiry int64 `long:"tipexpiry" Description:"Invoice expiry time in seconds"`

	ReconnectInterval int64 `long:"reconnectinterval" Description:"Reconnect interval to LND in seconds"`
	KeepAliveInterval int64 `long:"keepaliveinterval" Description:"Send a dummy request to LND to prevent timeouts "`

	LND *backends.LND `group:"LND" namespace:"lnd"`

	Mail *notifications.Mail `group:"Mail" namespace:"mail"`
}

var cfg config

var backend backends.Backend

func initConfig() {
	cfg = config{
		ConfigFile: path.Join(getDefaultDataDir(), defaultConfigFile),

		DataDir: getDefaultDataDir(),

		LogFile:  path.Join(getDefaultDataDir(), defaultLogFile),
		LogLevel: defaultLogLevel,

		DatabaseFile: path.Join(getDefaultDataDir(), defaultDatabaseFile),

		RESTHost:    defaultRESTHost,
		TLSCertFile: defaultTLSCertFile,
		TLSKeyFile:  defaultTLSKeyFile,

		AccessDomain: defaultAccessDomain,

		TipExpiry: defaultTipExpiry,

		ReconnectInterval: defaultReconnectInterval,
		KeepAliveInterval: defaultKeepaliveInterval,

		LND: &backends.LND{
			GRPCHost:     defaultLndGRPCHost,
			CertFile:     path.Join(getDefaultLndDir(), defaultLndCertFile),
			MacaroonFile: getDefaultMacaroon(),
		},

		Mail: &notifications.Mail{
			Recipient: defaultRecipient,
			Sender:    defaultSender,

			SMTPServer:   defaultSTMPServer,
			SMTPSSL:      defaultSTMPSSL,
			SMTPUser:     defaultSTMPUser,
			SMTPPassword: defaultSTMPPassword,
		},
	}

	// Ignore unknown flags the first time parsing command line flags to prevent showing the unknown flag error twice
	flags.NewParser(&cfg, flags.IgnoreUnknown).Parse()

	errFile := flags.IniParse(cfg.ConfigFile, &cfg)

	// Parse flags again to override config file
	_, err := flags.Parse(&cfg)

	// Default log level if parsing fails
	logLevel := logging.DEBUG

	switch strings.ToLower(cfg.LogLevel) {
	case "info":
		logLevel = logging.INFO

	case "warning":
		logLevel = logging.WARNING

	case "error":
		logLevel = logging.ERROR
	}

	// Create data directory
	var errDataDir error
	var dataDirCreated bool

	if _, err := os.Stat(getDefaultDataDir()); os.IsNotExist(err) {
		errDataDir = os.Mkdir(getDefaultDataDir(), 0700)

		dataDirCreated = true
	}

	errLogFile := initLogger(cfg.LogFile, logLevel)

	// Show error messages
	if err != nil {
		log.Error("Failed to parse command line flags")
	}

	if errDataDir != nil {
		log.Error("Could not create data directory")
		log.Debug("Data directory path: " + getDefaultDataDir())

	} else if dataDirCreated {
		log.Debug("Created data directory: " + getDefaultDataDir())
	}

	if errFile != nil {
		log.Warning("Failed to parse config file: " + fmt.Sprint(errFile))
	} else {
		log.Debug("Parsed config file: " + cfg.ConfigFile)
	}

	if errLogFile != nil {
		log.Error("Failed to initialize log file: " + fmt.Sprint(err))

	} else {
		log.Debug("Initialized log file: " + cfg.LogFile)
	}

	database.UseLogger(*log)
	backends.UseLogger(*log)
	notifications.UseLogger(*log)

	backend = cfg.LND
}

func getDefaultDataDir() (dir string) {
	homeDir := getHomeDir()

	switch runtime.GOOS {
	case "windows":
		fallthrough

	case "darwin":
		dir = path.Join(homeDir, defaultDataDir)

	default:
		dir = path.Join(homeDir, "."+strings.ToLower(defaultDataDir))
	}

	return cleanPath(dir)
}

// If the mainnet macaroon does exists it is preffered over all others
func getDefaultMacaroon() string {
	networksDir := filepath.Join(getDefaultLndDir(), "/data/chain/bitcoin/")
	mainnetMacaroon := filepath.Join(networksDir, "mainnet/", defaultMacaroonFile)

	if _, err := os.Stat(mainnetMacaroon); err == nil {
		return mainnetMacaroon
	}

	networks, err := ioutil.ReadDir(networksDir)

	if err == nil && len(networks) != 0 {
		for _, subDir := range networks {
			if subDir.IsDir() {
				return filepath.Join(networksDir, networks[0].Name(), defaultMacaroonFile)
			}
		}
	}

	log.Warning("Could not find macaroon file")
	return ""
}

func getDefaultLndDir() (dir string) {
	homeDir := getHomeDir()

	switch runtime.GOOS {
	case "darwin":
		fallthrough

	case "windows":
		dir = path.Join(homeDir, "Lnd")

	default:
		dir = path.Join(homeDir, ".lnd")
	}

	return cleanPath(dir)
}

func getHomeDir() (dir string) {
	usr, err := user.Current()

	if err == nil {
		switch runtime.GOOS {
		case "darwin":
			dir = path.Join(usr.HomeDir, "Library/Application Support")

		case "windows":
			dir = path.Join(usr.HomeDir, "AppData/Local")

		default:
			dir = usr.HomeDir
		}

	}

	return cleanPath(dir)
}

func cleanPath(path string) string {
	path = filepath.Clean(os.ExpandEnv(path))

	return strings.Replace(path, "\\", "/", -1)
}
