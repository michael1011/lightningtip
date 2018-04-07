package main

import (
	"fmt"
	"github.com/urfave/cli"
	"strconv"
	"time"
)

// TODO: add description?
// TODO: show sender of tips?
var summaryCommand = cli.Command{
	Name:   "summary",
	Usage:  "Shows a summary of received tips",
	Action: summary,
}

func summary(ctx *cli.Context) error {
	db, err := openDatabase(ctx)

	if err == nil {
		rows, err := db.Query("SELECT * FROM tips ORDER BY date DESC")

		if err == nil {
			var tips int64
			var sum int64

			var unixDate int64
			var amount int64
			var message string

			for rows.Next() {
				err = rows.Scan(&unixDate, &amount, &message)

				tips++
				sum += amount
			}

			if err == nil {
				date := time.Unix(unixDate, 0)

				fmt.Println("Received " + strconv.FormatInt(tips, 10) + " of tips since " + date.Format("02-01-2006") +
					" totalling " + strconv.FormatInt(sum, 10) + " satoshis")
			}

		}

	}

	return err
}
