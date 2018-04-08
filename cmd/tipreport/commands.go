package main

import (
	"database/sql"
	"fmt"
	"github.com/urfave/cli"
	"strconv"
	"time"
)

type tip struct {
	Date    string
	Amount  string
	Message string
}

// TODO: add description?
var summaryCommand = cli.Command{
	Name:   "summary",
	Usage:  "Shows a summary of received tips",
	Action: summary,
}

func summary(ctx *cli.Context) error {
	db, err := openDatabase(ctx)

	if err == nil {
		rows, err := getTips(db)

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
				date := formatUnixDate(unixDate)

				// Trim hours and minutes
				date = date[:len(date)-6]

				fmt.Println("Received " + formatInt(tips) + " tips since " + date +
					" totalling " + formatInt(sum) + " satoshis")
			}

		}

	}

	return err
}

// TODO: show sender of tips?
var listCommand = cli.Command{
	Name:   "list",
	Usage:  "Shows all received tips",
	Action: list,
}

func list(ctx *cli.Context) error {
	db, err := openDatabase(ctx)

	if err == nil {
		rows, err := getTips(db)

		if err == nil {
			var tips []tip

			// To ensure that the grid looks right
			maxAmountSize := 6

			var unixDate int64
			var amount int64
			var message string

			for rows.Next() {
				err = rows.Scan(&unixDate, &amount, &message)

				amountString := formatInt(amount)

				tips = append(tips, tip{
					Date:    formatUnixDate(unixDate),
					Amount:  amountString,
					Message: message,
				})

				if amountSize := len(amountString); amountSize > maxAmountSize {
					maxAmountSize = amountSize
				}

			}

			fmt.Println("Date              Amount" + getSpacing(6, maxAmountSize) + "Message")

			for _, tip := range tips {
				tipSpacing := getSpacing(len(tip.Amount), maxAmountSize)

				fmt.Println(tip.Date + "  " + tip.Amount + tipSpacing + tip.Message)
			}

		}

	}

	return err
}

func getSpacing(entrySize int, maxSize int) string {
	spacing := "  "

	spacingSize := maxSize - entrySize

	for spacingSize > 0 {
		spacing += " "

		spacingSize--
	}

	return spacing
}

func formatUnixDate(unixDate int64) string {
	date := time.Unix(unixDate, 0)

	return date.Format("02-01-2006 15:04")
}

func formatInt(i int64) string {
	return strconv.FormatInt(i, 10)
}

func getTips(db *sql.DB) (rows *sql.Rows, err error) {
	return db.Query("SELECT * FROM tips ORDER BY date DESC")
}
