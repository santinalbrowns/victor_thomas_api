package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func Init() error {
	//database credentials
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PSWD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	// Create a DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", user, password, host, port, dbname)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("could not connect to the database: %w", err)
	}

	// Test the database connection
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("could not ping the database: %w", err)
	}

	return nil
}

func Close() error {
	if DB != nil {
		err := DB.Close()
		if err != nil {
			return fmt.Errorf("could not close the database connection: %w", err)
		}
	}
	return nil
}
