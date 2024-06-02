package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"time"

	"xelbot.com/auto-notes/server/internal/application"
	"xelbot.com/auto-notes/server/internal/service"
	"xelbot.com/auto-notes/server/proto"
)

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

	appContainer := application.Container{
		InfoLog:  infoLog,
		ErrorLog: errorLog,
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

func handleError(err error, logger *log.Logger) {
	if err != nil {
		logger.Println(fmt.Sprintf("%s", debug.Stack()))
		logger.Fatal(err)
	}
}
