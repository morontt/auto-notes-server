package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/kataras/jwt"
	"xelbot.com/auto-notes/server/internal/application"
	"xelbot.com/auto-notes/server/internal/middlewares"
	"xelbot.com/auto-notes/server/internal/services/auth"
	"xelbot.com/auto-notes/server/internal/services/server"
	pbAuth "xelbot.com/auto-notes/server/rpc/auth"
	pbServer "xelbot.com/auto-notes/server/rpc/server"
)

func init() {
	jwt.Unmarshal = jwt.UnmarshalWithRequired
}

func main() {
	var err error
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	var cli struct {
		ConfigFile string `default:"config/config.toml" help:"Path to config file"`
	}

	kong.Parse(&cli, kong.Description("gRPC server for autonotes app."))

	handleError(application.LoadConfig(cli.ConfigFile), errorLog)
	cnf := application.GetConfig()

	var logLevel = new(slog.LevelVar)
	switch cnf.LogLevel {
	case "debug":
		logLevel.Set(slog.LevelDebug)
	}

	infoLog := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	infoLog.Debug("Loading configuration", "log_level", cnf.LogLevel, "time_zone", cnf.TimeZone)

	appContainer := application.Container{}
	appContainer.SetLogger(infoLog)

	handleError(appContainer.SetupDatabase(), errorLog)

	authImpl := auth.NewAuthService(appContainer)
	authHandler := pbAuth.NewAuthServer(authImpl)

	userRepoImpl := server.NewUserRepositoryService(appContainer)
	userRepoHandler := pbServer.NewUserRepositoryServer(userRepoImpl)

	fuelRepoImpl := server.NewFuelRepositoryService(appContainer)
	fuelRepoHandler := pbServer.NewFuelRepositoryServer(fuelRepoImpl)

	mux := http.NewServeMux()
	mux.Handle(authHandler.PathPrefix(), authHandler)
	mux.Handle(userRepoHandler.PathPrefix(), middlewares.WithAuthorization(appContainer, userRepoHandler))
	mux.Handle(fuelRepoHandler.PathPrefix(), middlewares.WithAuthorization(appContainer, fuelRepoHandler))

	handler := middlewares.Clacks().Middleware(mux)
	handler = middlewares.RequestID(handler)

	server := &http.Server{
		Handler:      handler,
		Addr:         ":" + strconv.Itoa(cnf.Port),
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	go func() {
		infoLog.Info("Starting server", "port", cnf.Port)
		err = server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			handleError(err, errorLog)
		}
	}()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	<-exit

	infoLog.Info("Shutting down...")
	err = server.Shutdown(context.Background())
	if err != nil {
		errorLog.Fatal(err)
	}
	infoLog.Info("Server stopped")

	err = appContainer.Stop()
	if err != nil {
		errorLog.Fatal(err)
	}
	infoLog.Info("Application stopped")
}

func handleError(err error, logger *log.Logger) {
	if err != nil {
		logger.Printf("%s\n", debug.Stack())
		logger.Fatal(err)
	}
}
