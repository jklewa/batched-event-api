package handler

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/jklewa/batched-event-api/api/types"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type UserEventHandler struct {
	sync.Mutex
	outputDir      string
	batchInterval  time.Duration
	autoCloseAfter time.Duration

	openTime       time.Time
	firstEventTime time.Time
	openFile       *os.File
	openDataWriter *csv.Writer
}

func NewUserEventHandler(outputDir string, batchInterval time.Duration, autoCloseAfter time.Duration) *UserEventHandler {
	handler := &UserEventHandler{
		outputDir:      outputDir,
		batchInterval:  batchInterval,
		autoCloseAfter: autoCloseAfter,
		openTime:       time.Time{},
		firstEventTime: time.Time{},
		openFile:       nil,
		openDataWriter: nil,
	}
	log.SetOutput(os.Stderr)
	if autoCloseAfter > 0 {
		go handler.closeOpenFileWriterRoutine()
	}
	return handler
}

func (h *UserEventHandler) closeOpenFileWriterRoutine() {
	for {
		time.Sleep(h.autoCloseAfter)
		h.closeExpiredFile()
	}
}

func (h *UserEventHandler) closeExpiredFile() {
	if h.openFile != nil && time.Since(h.openTime) >= h.autoCloseAfter {
		log.Printf("closing expired file: %s\n", h.openFile.Name())
		err := h.CloseFileAndWriter()
		if err != nil {
			log.Fatalf("Unable to close expired file: %s %v\n", h.openFile.Name(), err)
		}
	}
}

func (h *UserEventHandler) Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	scanner := bufio.NewScanner(r.Body)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			log.Printf("Error closing request body: %s\n", err)
		}
	}()

	err := h.handleUserEvent(w, scanner)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error handling request: %s", err.Error()),
			http.StatusInternalServerError)
		log.Printf("Error handling request: %s\n", err)
	}
}

func (h *UserEventHandler) CloseFileAndWriter() error {
	h.Lock()
	defer func() {
		h.openTime = time.Time{}
		h.firstEventTime = time.Time{}
		h.openDataWriter = nil
		h.openFile = nil
		h.Unlock()
	}()
	if h.openDataWriter != nil {
		h.openDataWriter.Flush()
		if err := h.openDataWriter.Error(); err != nil {
			return err
		}
	}
	if h.openFile != nil {
		if err := h.openFile.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (h *UserEventHandler) createNewFileWriter(event *types.UserEvent) error {
	h.Lock()
	defer h.Unlock()
	newOutputFileName := fmt.Sprintf(
		"user-events-%s.csv",
		event.Time.Format("20060102-150405"))
	fullPath := filepath.Join(h.outputDir, newOutputFileName)
	var err error
	if _, err = os.Stat(fullPath); !os.IsNotExist(err) {
		return fmt.Errorf("file already exists: %s", fullPath)
	}
	if h.openFile, err = os.Create(fullPath); err != nil {
		return fmt.Errorf("failed to create new file: %s %v", fullPath, err)
	}
	h.openDataWriter = csv.NewWriter(h.openFile)
	h.openTime = time.Now()
	h.firstEventTime = event.Time
	return nil
}

func (h *UserEventHandler) writeEventDataCSV(event *types.UserEvent) error {
	err := h.openDataWriter.Write(event.CSVData())
	if err != nil {
		return fmt.Errorf("error writing event data: %s", err)
	}
	return nil
}

func (h *UserEventHandler) handleUserEvent(w http.ResponseWriter, scanner *bufio.Scanner) error {
	for scanner.Scan() {
		// Parse a new line of JSON user event data
		var event types.UserEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return err
		}

		// Do we need to rotate output files?
		if event.Time.Sub(h.firstEventTime) >= h.batchInterval {
			if h.openDataWriter != nil {
				err := h.CloseFileAndWriter()
				if err != nil {
					return err
				}
			}
		}

		// Do we need to start a new output file?
		if h.openDataWriter == nil {
			err := h.createNewFileWriter(&event)
			if err != nil {
				return err
			}
		}

		err := h.writeEventDataCSV(&event)
		if err != nil {
			return err
		}
	}

	// We've processed all the data, respond and start cleanup
	w.WriteHeader(http.StatusOK)
	_, err := fmt.Fprint(w, "OK")
	if err != nil {
		return err
	}
	log.Println("Response sent")

	if h.openDataWriter != nil {
		h.openDataWriter.Flush()
		// Our closeOpenFileWriterRoutine will eventually close it
	}
	return nil
}
