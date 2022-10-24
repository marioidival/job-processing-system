package main

import (
	"context"
	"errors"
	"flag"
	"github.com/marioidival/job-processing-system/cmd/api/handlers"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/peterbourgon/ff"

	"github.com/marioidival/job-processing-system/pkg/database"
)

func main() {

	if err := run(); err != nil {
		log.Fatal(err.Error())
	}
}

func run() error {
	fs := flag.NewFlagSet("api", flag.ExitOnError)

	var databaseURL string

	fs.StringVar(&databaseURL, "database-url", "", "e.g., postgres://username:password@localhost:5432/database_name")

	err := ff.Parse(fs, os.Args[1:], ff.WithEnvVarNoPrefix())
	if err != nil {
		return err
	}
	ctx := context.Background()
	dbc, err := database.Open(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer dbc.Close()

	addr := ":3000"

	mux := handlers.SetupHandlers(dbc)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	serverErrors := make(chan error, 1)
	go func() {
		log.Println(ctx, "startup job processing system api", "PORT", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case serverError := <-serverErrors:
		return errors.Unwrap(serverError)

	case sig := <-quit:
		log.Println("Server is shutting down", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if shutdownErr := server.Shutdown(ctx); shutdownErr != nil {
			defer func() {
				closeErr := server.Close()
				if closeErr != nil {
					log.Fatalln("Could not close server", closeErr)
				}
			}()
			log.Fatalln("Could not gracefully shutdown the server")
		}
		close(done)
	case <-done:
		return nil
	}

	return nil
}
