# Define the first stage with the name 'builder'
FROM golang:1.24-alpine AS builder

# Set sqlite
ENV CGO_ENABLED=1
RUN apk add \
    # Important: required for go-sqlite3
    gcc \
    # Required for Alpine
    musl-dev

# Set the working directory inside the container
WORKDIR /collector

## Кэшируем зависимости Go, чтобы ускорить повторные билды
#ENV GOPROXY=https://proxy.golang.org
#RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked true

# Copy go.mod and go.sum to the working directory
COPY go.mod ./

# Download Go module dependencies
RUN go mod download

# Copy the source code to the working directory
# COPY /main.go main.go
COPY . .

# Attempt to build the Go application and capture output
RUN go build -o app ./cmd/collector/main.go

# Define the second stage from which the final image will be created
FROM alpine:latest

# Copy the compiled application from the first stage to the final image
COPY --from=builder /collector/app /app

ENTRYPOINT ["/app"]
CMD ["collect"]
