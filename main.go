package main

import (
	"flag"
	"fmt"
	"github.com/jklewa/batched-event-api/api/handler"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	outputDir := flag.String("o", "./data", "Output directory")
	hostName := flag.String("host", "localhost", "Hostname")
	batchInterval := 5 * time.Minute
	autoCloseAfter := 15 * time.Second // this is short for convenience - should be longer
	flag.Parse()
	validDir(*outputDir)

	http.HandleFunc("/user/event", handler.NewUserEventHandler(*outputDir, batchInterval, autoCloseAfter).Handler)

	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:8080", *hostName), nil))
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
