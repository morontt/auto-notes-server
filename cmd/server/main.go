package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"xelbot.com/auto-notes/server/internal/app"
	"xelbot.com/auto-notes/server/internal/service"
	"xelbot.com/auto-notes/server/proto"
)

func main() {
	infoLog := slog.New(slog.NewTextHandler(os.Stdout, nil))
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	appContainer := app.Container{
		InfoLog:  infoLog,
		ErrorLog: errorLog,
	}

	authImpl := service.NewAuthService(appContainer)
	authHandler := proto.NewAuthServer(authImpl)

	mux := http.NewServeMux()
	mux.Handle(authHandler.PathPrefix(), authHandler)

	server := &http.Server{
		Handler:      mux,
		Addr:         ":8080",
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	infoLog.Info("Starting server on 8080 port")
	handleError(server.ListenAndServe(), errorLog)
}

func handleError(err error, logger *log.Logger) {
	if err != nil {
		logger.Println(fmt.Sprintf("%s", debug.Stack()))
		logger.Fatal(err)
	}
}
