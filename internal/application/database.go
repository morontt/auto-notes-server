package application

import (
	"database/sql"
	"log/slog"
	"time"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	"github.com/go-sql-driver/mysql"
)

func getDBConnection(logger *slog.Logger) (db *sql.DB, err error) {
	var i int
	var location *time.Location

	location, err = time.LoadLocation(cfg.TimeZone)
	if err != nil {
		return
	}

	goqu.SetTimeLocation(location)
	dsn := createDSN(location)

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

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func createDSN(loc *time.Location) string {
	dbConfig := mysql.NewConfig()
	dbConfig.User = cfg.Database.User
	dbConfig.Passwd = cfg.Database.Password
	dbConfig.Addr = cfg.Database.Host + ":3306"
	dbConfig.DBName = cfg.Database.Name
	dbConfig.Loc = loc
	dbConfig.ParseTime = true

	return dbConfig.FormatDSN()
}
