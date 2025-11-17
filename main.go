package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"congrats-project.com/config"
	"congrats-project.com/db"
	"congrats-project.com/routes"
	"congrats-project.com/scheduler"
	"github.com/joho/godotenv"
)

/*
Future Improvements:
	- Find a way to remove inactive events within the setup scheduler. this will be a DB query (added logic to event.go, just need to add safeguards in case email doesn't send)

*/

func init() {
	// Check for GO_ENV
	env := os.Getenv("GO_ENV")

	if env == "" || env == "development" {
		// Load .env file if in development mode
		if err := godotenv.Load(); err != nil {
			log.Println("Warning: Error loading .env file:", err)
		} else {
			log.Printf("Development mode: Loaded .env file")
		}
	} else {
		log.Printf("Running in %s mode: Using system environment variables\n", env)
	}

	// Load web configurations
	config.LoadConfigs()

	// Initialize database
	db.InitDB()

	// Start scheduler for daily email notifications
	// Runs at 8AM system time (configurable)
	scheduler.StartScheduler(8)

	// Start cleanup scheduler with 30-day grace period
	// Runs weekly to delete events in which email was sent 30+ days ago
	scheduler.StartCleanupScheduler(30)
}

func main() {

	defer db.DB.Close()

	mux := routes.RegisterRoutes()

	s := &http.Server{
		Addr:           ":" + config.App.ServerPort,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Println("Server starting on: " + config.App.BaseURL)
	log.Fatal(s.ListenAndServe())
}
