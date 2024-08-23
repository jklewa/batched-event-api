package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/jklewa/batched-event-api/api/handler"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	outputDir := flag.String("o", "./data", "Output directory")
	hostName := flag.String("host", "localhost", "Hostname")
	batchInterval := 5 * time.Minute
	// autoCloseAfter := 10 * time.Minute // 10 mins after last write
	autoCloseAfter := 15 * time.Second
	// autoCloseInterval := 1 * time.Minute
	autoCloseInterval := 1 * time.Second
	flag.Parse()
	validDir(*outputDir)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:8080", *hostName),
		Handler: nil, // uses default mux
	}

	eventHandler := handler.NewUserEventHandler(*outputDir, batchInterval, autoCloseAfter, autoCloseInterval)
	http.HandleFunc("/user/event", eventHandler.Handler)

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
		log.Println("Stopped serving new connections.")
	}()

	// 1. Catch my signal interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		log.Printf("graceful shutdown")
		err := eventHandler.Shutdown()
		if err != nil {
			log.Fatalf("Error during shutdown: %s", err.Error())
		}
		log.Println("Graceful shutdown complete.")
		shutdownRelease()
	}()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP shutdown error: %v", err)
	}
}

func validDir(path string) {
	// Verify that path exists and is a valid directory
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		log.Fatalf("Output directory %s does not exist.", path)
	} else if err != nil {
		log.Fatalf("Error checking output directory: %v", err)
	} else if !info.IsDir() {
		log.Fatalf("Output path %s is not a directory.", path)
	}
}
