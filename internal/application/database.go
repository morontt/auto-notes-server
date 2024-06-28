package application

import (
	"database/sql"
	"log/slog"
	"time"

	"github.com/go-sql-driver/mysql"
)

func getDBConnection(logger *slog.Logger) (db *sql.DB, err error) {
	var i int

	dsn, err := createDSN()
	if err != nil {
		return
	}

	for i < 5 {
		logger.Info("Trying to connect to the database")
		db, err = openDB(dsn)
		if err == nil {
			logger.Info("The database is connected")

			return
		} else {
			logger.Error(err.Error())
		}

		i++
		time.Sleep(1000 * time.Millisecond)
	}

	return nil, err
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func createDSN() (string, error) {
	location, err := time.LoadLocation(cfg.TimeZone)
	if err != nil {
		return "", err
	}

	dbConfig := mysql.NewConfig()
	dbConfig.User = cfg.Database.User
	dbConfig.Passwd = cfg.Database.Password
	dbConfig.Addr = cfg.Database.Host + ":3306"
	dbConfig.DBName = cfg.Database.Name
	dbConfig.Loc = location
	dbConfig.ParseTime = true

	return dbConfig.FormatDSN(), nil
}
