package main

import (
	"context"
	"fmt"
	"github.com/hunttraitor/dialed-in-backend/application"
	"github.com/joho/godotenv"
	"os"
	"os/signal"
)

func main() {
	// Load env variables
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Could not load .env file:", err)
	}

	// Create new instance of an application
	app, err := application.New()
	if err != nil {
		fmt.Println("Could not build application:", err)
	}
	// Close the db on shutdown
	defer app.Close()

	// Create a context that gracefully shutdowns on a keyboard interrupt
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Start the application
	err = app.Start(ctx)
	if err != nil {
		fmt.Println("Could not start server:", err)
	}
}
