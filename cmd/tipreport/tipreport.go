package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/urfave/cli"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	version = "1.0.0"

	defaultDataDir      = "LightningTip"
	defaultDatabaseFile = "tips.db"
)

func main() {
	app := cli.NewApp()

	app.Name = "tipreport"
	app.Usage = "display received tips"

	app.Version = version

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "databasefile",
			Value: getDefaultDatabaseFile(),
			Usage: "path to database file",
		},
	}

	app.Commands = []cli.Command{
		summaryCommand,
		listCommand,
	}

	err := app.Run(os.Args)

	if err != nil {
		fmt.Println(err)
	}

}

func openDatabase(ctx *cli.Context) (db *sql.DB, err error) {
	db, err = sql.Open("sqlite3", ctx.GlobalString("databasefile"))

	return db, err
}

func getDefaultDatabaseFile() (dir string) {
	usr, _ := user.Current()

	switch runtime.GOOS {
	case "darwin":
		dir = path.Join(usr.HomeDir, "Library/Application Support", defaultDataDir, defaultDatabaseFile)

	case "windows":
		dir = path.Join(usr.HomeDir, "AppData/Local", defaultDataDir, defaultDatabaseFile)

	default:
		dir = path.Join(usr.HomeDir, "."+strings.ToLower(defaultDataDir), defaultDatabaseFile)
	}

	return cleanPath(dir)
}

func cleanPath(path string) string {
	path = filepath.Clean(os.ExpandEnv(path))

	return strings.Replace(path, "\\", "/", -1)
}
