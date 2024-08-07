package main

import (
	"flag"
	"fmt"
	"github.com/jklewa/batched-event-api/api/handler"
	"log"
	"net/http"
	"time"
)

func main() {
	outputDir := flag.String("o", "/data", "Output directory")
	hostName := flag.String("h", "", "Hostname")
	flag.Parse()
	batchInterval := 5 * time.Minute
	autoCloseAfter := 15 * time.Second

	http.HandleFunc("/user/event", handler.NewUserEventHandler(*outputDir, batchInterval, autoCloseAfter).Handler)

	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:8080", *hostName), nil))
}
