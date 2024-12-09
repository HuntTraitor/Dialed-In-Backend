# Build stage for development environment
FROM golang:alpine AS dev

# Add Maintainer info
LABEL maintainer="Hunter Tratar"

# Setup folders
RUN mkdir /app
WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

RUN apk add --no-cache make gcc libc-dev git

# Setup hot-reload for dev stage
RUN go install -mod=mod github.com/githubnemo/CompileDaemon
RUN go get -v golang.org/x/tools/gopls
RUN go install github.com/pressly/goose/v3/cmd/goose@latest
COPY Makefile .

ENV PATH=$PATH:/root/go/bin

# Entry point for dev environment
ENTRYPOINT CompileDaemon --build="make build" --command="./bin/api -smtp-host=localhost -smtp-port=1025 -smtp-username= -smtp-password= -metrics=true"

# Build stage for production environment
FROM golang:alpine AS prod

# Add Maintainer info
LABEL maintainer="Hunter Tratar"

# Setup folders
RUN mkdir /app
WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Install necessary build tools for production
RUN apk add --no-cache make gcc libc-dev git

# Copy the source code and build the application
COPY . .

# Build the application in production
RUN make build

# Final image for production, copying over the built binary
FROM alpine:latest

# Setup the app folder
WORKDIR /app

# Copy the compiled binary from the prod build stage
COPY --from=prod /app/bin/linux/_amd64/api /app/bin/

# Expose the production app port
EXPOSE 3000

# Ensure the binary is executable
RUN chmod +x /app/bin/api

# Command to run the production application
ENTRYPOINT ["/app/bin/api", "-metrics=true", "-env=production"]
