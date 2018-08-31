package database

import (
	"database/sql"
	"fmt"
	"time"

	// The sqlite drivers have to be imported to establish a connection to the database
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// InitDatabase is initializing the database
func InitDatabase(databaseFile string) (err error) {
	db, err = sql.Open("sqlite3", databaseFile)

	if err == nil {
		db.Exec("CREATE TABLE IF NOT EXISTS `tips` (`date` INTEGER, `amount` INTEGER, `message` VARCHAR)")
	}

	return err
}

// AddSettledInvoice is adding a settled invoice to the database
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
