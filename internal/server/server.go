
package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"backend/internal/database" // Import your database package
)

type Server struct {
	port int
	db database.Service // Now of type database.Service
}

// NewServer now takes the database.Service as an argument
func NewServer(dbService database.Service) *http.Server { // Changed signature
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	if port == 0 {
		port = 8080 // Default port if not set
	}

	// Create an instance of your Server struct
	serverInstance := &Server{
		port: port,
		db:   dbService, // Assign the passed dbService
	}

	// Declare http.Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", serverInstance.port),
		Handler:      serverInstance.RegisterRoutes(), // This calls your existing RegisterRoutes
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}

