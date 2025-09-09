FROM golang:1.21-alpine

# Install git (needed to fetch modules)
RUN apk add --no-cache git

WORKDIR /app

# Copy go.mod and go.sum first
COPY go.mod go.sum ./

# Download all modules, including fsnotify
RUN go mod download
# Download modules including transitive ones
RUN go get github.com/fsnotify/fsnotify
# Copy all source files
COPY *.go ./

# Build the binary including all files
RUN go build -o wal_app *.go

# Create data directory
RUN mkdir /data
VOLUME /data

CMD ["./wal_app"]

