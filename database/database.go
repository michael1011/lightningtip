package database

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

var db *sql.DB

func InitDatabase(databaseFile string) (err error) {
	db, err = sql.Open("sqlite3", databaseFile)

	if err == nil {
		db.Exec("CREATE TABLE IF NOT EXISTS `tips` (`date` INTEGER, `amount` INTEGER, `message` VARCHAR)")
	}

	return err
}

func AddSettledInvoice(amount int64, message string) {
	stmt, err := db.Prepare("INSERT INTO tips(date, amount, message) values(?, ?, ?)")

	if err == nil {
		_, err = stmt.Exec(time.Now().Unix(), amount, message)
	}

	if err == nil {
		stmt.Close()

	} else {
		log.Error("Could not insert into database: " + fmt.Sprint(err))
	}

}
