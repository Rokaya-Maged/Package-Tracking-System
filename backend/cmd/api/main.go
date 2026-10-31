package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"tracking-system/config"
	"tracking-system/internal/router"
)

func main() {
	if err := config.ConnectDB(); err != nil {
		log.Fatal("database connection failed: ", err)
	}
	defer config.DB.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router.New(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("API server listening on :%s", port)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("API server failed: ", err)
	}
}
