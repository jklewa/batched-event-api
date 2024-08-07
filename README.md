# User Event Webservice

This project consists of a web service designed to handle incoming user event data.

The service receives newline-delimited JSON via HTTP POST requests, converts the
data into CSV format, and writes the data to disk in 5-minute intervals based on
the timestamp of each event.

See `./api/handler/userevent.go`

### Input
The web service:
- Receives a `POST` HTTP request with a payload containing User Event data.
- The payload is newline-delimited JSON.
- Rows are ordered by the `time` field across requests.
- Only one request is handled at a time (requests are synchronous).

### Output
The web service:
- Responds with a `200 OK` status after reading the request payload.
- Converts received JSON data to CSV and writes it to disk.
- Batches data into 5-minute intervals according to the `time` field.
- Names files based on the timestamp of the first data point.
- Allows multiple requests to contribute to the same 5-minute interval files.
- Closes and commits files once they go beyond the 5-minute interval.

### Examples
Two `POST /user/event` requests are made:
1. The first payload contains events starting at `2024-07-01T02:03:04Z` with the last event at `2024-07-01T02:11:05Z`, occurring every second.
2. The second payload contains events starting at `2024-07-01T02:12:06Z` with the last event at `2024-07-01T02:15:07Z`, occurring every second.

This results in 3 CSV files:
- The first CSV contains data from `2024-07-01T02:03:04Z` to `2024-07-01T02:08:04Z`.
- The second CSV contains data from `2024-07-01T02:08:04Z` to `2024-07-01T02:13:04Z`.
- The third CSV contains data from `2024-07-01T02:13:04Z` to `2024-07-01T02:18:04Z`.

See `./api/handler/userevent_test.go`

## Usage

Docker Compose
```bash
# Run the web service
docker compose up
```

Locally
```bash
# Run the web service
go run main.go -o ./data
```

Use a tool like `curl` or `Postman` to send `POST` requests with the required NDJSON payload to the web service endpoint.

```bash
curl --request POST \
  --url http://localhost:8080/user/event \
  --header 'Content-Type: application/x-ndjson' \
  --data-binary @data/initial-dataset-01.ndjson
```

## Testing

Tests covering `api/handler/userevent.go` are in `api/handler/userevent_test.go` and can be run with `go test`.

```bash
go test ./...

> ok  	github.com/jklewa/batched-event-api/api/handler	0.201s
```