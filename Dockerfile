# Use the official Golang image to create a build artifact.
# https://hub.docker.com/_/golang
FROM golang:1.22.5 AS builder

# Set necessary environment variables for Go.
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Copy the local package to the container's workspace.
WORKDIR /app
COPY go.mod ./

# Download all the dependencies.
RUN go mod download
COPY . .

# Build the project.
RUN go build -o /bin/main

# Use a minimal alpine image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/
# Copy the binary from the builder.
COPY --from=builder /bin/main /bin/main

# Expose port 8080
EXPOSE 8080

# Run the binary.
CMD ["/bin/main"]