# Use ARG for build-time variables
ARG GO_VERSION=1.22.4
FROM golang:${GO_VERSION}-alpine

# Install PostgreSQL client
RUN apk add --no-cache postgresql-client

WORKDIR /build

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o human-network-iam ./cmd

EXPOSE 8080

# The command to run the application
CMD ["./human-network-iam"]
