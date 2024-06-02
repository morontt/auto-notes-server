package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/kataras/jwt"
	"xelbot.com/auto-notes/server/internal/application"
	"xelbot.com/auto-notes/server/internal/service"
	"xelbot.com/auto-notes/server/proto"
)

func init() {
	jwt.Unmarshal = jwt.UnmarshalWithRequired
}

func main() {
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	handleError(application.LoadConfig(), errorLog)
	cnf := application.GetConfig()

	var logLevel = new(slog.LevelVar)
	switch cnf.LogLevel {
	case "debug":
		logLevel.Set(slog.LevelDebug)
	}

	infoLog := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	infoLog.Debug("Loading configuration", "config", cnf)

	db, err := getDBConnection(cnf, infoLog)
	if err != nil {
		errorLog.Fatal(err)
	}

	appContainer := application.Container{
		InfoLog:  infoLog,
		ErrorLog: errorLog,
		DB:       db,
	}

	authImpl := service.NewAuthService(appContainer)
	authHandler := proto.NewAuthServer(authImpl)

	mux := http.NewServeMux()
	mux.Handle(authHandler.PathPrefix(), authHandler)

	server := &http.Server{
		Handler:      mux,
		Addr:         ":" + strconv.Itoa(cnf.Port),
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	infoLog.Info("Starting server", "port", cnf.Port)
	handleError(server.ListenAndServe(), errorLog)
}

func getDBConnection(cnf application.Config, logger *slog.Logger) (db *sql.DB, err error) {
	var i int

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:3306)/%s?parseTime=true&loc=Europe%%2fMoscow",
		cnf.Database.User,
		cnf.Database.Password,
		cnf.Database.Host,
		cnf.Database.Name,
	)

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

func handleError(err error, logger *log.Logger) {
	if err != nil {
		logger.Println(fmt.Sprintf("%s", debug.Stack()))
		logger.Fatal(err)
	}
}
