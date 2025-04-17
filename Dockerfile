# Define the first stage with the name 'builder'
FROM golang:1.24-alpine AS builder

# Set the working directory inside the container
WORKDIR /workspace

# Copy go.mod and go.sum to the working directory
COPY /go.mod ./go.sum ./

# Download Go module dependencies
RUN go mod download

# Copy the source code to the working directory
COPY . .

# Attempt to build the Go application and capture output
RUN go build -o app ./main.go

# Define the second stage from which the final image will be created
FROM alpine:latest

# Copy the compiled application from the first stage to the final image
COPY --from=builder /workspace/app /app

# Expose the port that the application will run on
EXPOSE 8080
ENTRYPOINT ["/app"]
