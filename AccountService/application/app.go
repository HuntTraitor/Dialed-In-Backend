package application

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"os"
	"time"
)

type App struct {
	router http.Handler
	db     *pgxpool.Pool
}

func New() (*App, error) {
	// Grab database url form environment variable
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	// Establish a connection to that database
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, fmt.Errorf("could not connect to database: %w", err)
	}

	// Initialize the App struct
	app := &App{
		db: pool,
	}
	app.loadRoutes()
	return app, nil
}

func (a *App) Start(ctx context.Context) error {
	// Create a new server struct
	server := http.Server{
		Addr:    ":3000",
		Handler: a.router,
	}

	ch := make(chan error, 1)

	// Run the server
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("could not start server: %w", err)
		}
		close(ch)
	}()

	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()
		return server.Shutdown(timeout)
	}
}

func (a *App) Close() {
	// Close the database connection
	a.db.Close()
}
