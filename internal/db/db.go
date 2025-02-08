package db

import (
    "database/sql"
    "fmt"

    _ "github.com/go-sql-driver/mysql" // or your preferred driver
)

// NewMySQLClient opens a connection to the database.
func NewMySQLClient(user, pass, host, dbname string) (*sql.DB, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", user, pass, host, dbname)
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, err
    }
    // ping the DB to ensure connectivity:
    if err = db.Ping(); err != nil {
        return nil, err
    }
    return db, nil
}
