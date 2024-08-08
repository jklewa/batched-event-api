package handler

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"github.com/jklewa/batched-event-api/api/types"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func generateUserEvents(start time.Time, interval time.Duration, end time.Time) []types.UserEvent {
	num := int(end.Sub(start) / interval)
	events := make([]types.UserEvent, num)
	for i := 0; i < num; i++ {
		events[i] = types.UserEvent{
			Time: start.Add(time.Duration(i) * interval),
		}
	}
	return events
}

func Test_userEventHandler_handleUserEvent(t *testing.T) {
	/*
		Two `POST`s are made:
		* The first payload contains user event rows timestamped starting at `2024-07-01T02:03:04Z`,
		  with events occurring every second, with the last event at `2024-07-01T02:11:05`.
		* The second payload contains user event rows timestamped starting at `2024-07-01T02:12:06`,
		  with events occurring every second, with the last event at `2024-07-01T02:15:07`.

		The result should be 3 CSV files:
		* The first CSV should have all data that is [`2024-07-01T02:03:04Z`, `2024-07-01T02:08:04Z`).
		* The second CSV should have all data that is [`2024-07-01T02:08:04Z`, `2024-07-01T02:13:04Z`).
		* The third CSV should have all data that is [`2024-07-01T02:13:04Z`, `2024-07-01T02:18:04Z`).
	*/
	// Define the start times and intervals for the three batches of user events
	firstBatchStartTime := time.Date(2024, 7, 1, 2, 3, 4, 0, time.UTC)
	firstBatchEndTime := time.Date(2024, 7, 1, 2, 11, 5, 0, time.UTC)
	secondBatchStartTime := time.Date(2024, 7, 1, 2, 12, 6, 0, time.UTC)
	secondBatchEndTime := time.Date(2024, 7, 1, 2, 15, 7, 0, time.UTC)
	eventInterval := 1 * time.Second
	batchInterval := 5 * time.Minute
	fileIntervals := []time.Time{
		firstBatchStartTime.Add(1 * batchInterval),
		firstBatchStartTime.Add(2 * batchInterval),
		firstBatchStartTime.Add(3 * batchInterval),
	}
	wantEvents := map[string][]types.UserEvent{
		"user-events-20240701-020304.csv": generateUserEvents(firstBatchStartTime, eventInterval, fileIntervals[0]),
		"user-events-20240701-020804.csv": append(
			generateUserEvents(fileIntervals[0], eventInterval, firstBatchEndTime),
			generateUserEvents(secondBatchStartTime, eventInterval, fileIntervals[1])...,
		),
		"user-events-20240701-021304.csv": generateUserEvents(fileIntervals[1], eventInterval, secondBatchEndTime),
	}
	// Create a new temp directory and defer its deletion
	tmpDir, err := os.MkdirTemp("", "user-event-handler-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Fatalf("Failed to remove temp directory: %v", err)
		}
	}(tmpDir)
	handler := NewUserEventHandler(tmpDir, batchInterval, -1)

	// Marshal the []types.UserEvent into newline delimited json files
	batch1Buffer := new(bytes.Buffer)
	for _, event := range generateUserEvents(firstBatchStartTime, eventInterval, firstBatchEndTime) {
		encodedEvent, _ := json.Marshal(event)
		encodedEvent = append(encodedEvent, []byte("\n")...)
		batch1Buffer.Write(encodedEvent)
	}
	batch2Buffer := new(bytes.Buffer)
	for _, event := range generateUserEvents(secondBatchStartTime, eventInterval, secondBatchEndTime) {
		encodedEvent, _ := json.Marshal(event)
		encodedEvent = append(encodedEvent, []byte("\n")...)
		batch2Buffer.Write(encodedEvent)
	}

	req1, _ := http.NewRequest("POST", "/user/event", batch1Buffer)
	req2, _ := http.NewRequest("POST", "/user/event", batch2Buffer)
	w1 := httptest.NewRecorder()
	w2 := httptest.NewRecorder()
	handler.Handler(w1, req1)
	handler.Handler(w2, req2)
	resp1 := w1.Result()
	resp2 := w2.Result()
	if resp1.StatusCode != http.StatusOK {
		t.Errorf("Response 1 status: got %v want %v",
			resp1.StatusCode, http.StatusOK)
	}
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("Response 2 status: got %v want %v",
			resp2.StatusCode, http.StatusOK)
	}

	// Trigger a cleanup for any open files
	err = handler.CloseFileAndWriter()
	if err != nil {
		t.Fatalf("Failed to close file: %v", err)
	}

	gotEvents := map[string][]types.UserEvent{}
	fileInfos, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read directory %v: %v", tmpDir, err)
	}
	for _, fileInfo := range fileInfos {
		filePath := filepath.Join(tmpDir, fileInfo.Name())
		file, err := os.Open(filePath)
		if err != nil {
			t.Fatalf("Failed to open file %v: %v", filePath, err)
		}
		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil {
			t.Fatalf("Failed to read CSV file %v: %v", filePath, err)
		}
		for _, record := range records {
			eventTime, err := time.Parse(time.RFC3339Nano, record[0])
			if err != nil {
				t.Fatalf("Failed to parse timestamp in file %v: %v", filePath, err)
			}
			gotEvents[fileInfo.Name()] = append(gotEvents[fileInfo.Name()], types.UserEvent{
				Time: eventTime,
			})
		}
		if err = file.Close(); err != nil {
			t.Fatalf("Failed to close file %v: %v", filePath, err)
		}
	}

	// Compare wantEvents to gotEvents
	for fileName, want := range wantEvents {
		got, ok := gotEvents[fileName]
		if !ok {
			t.Errorf("File %s not found", fileName)
			continue
		}
		if len(got) != len(want) {
			t.Errorf("File %s: got %d events, want %d events", fileName, len(got), len(want))
			continue
		}
		for i := range got {
			if !got[i].Time.Equal(want[i].Time) {
				t.Errorf("File %s: event %d: got time %s, want time %s", fileName, i, got[i].Time, want[i].Time)
			}
		}
	}
}
