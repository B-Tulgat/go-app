FROM golang:1.18

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY main.go .
COPY .env .
# Build the Go app
RUN go build -o main .

EXPOSE 3729

# Run the Go app
CMD ["./main"]

