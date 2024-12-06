# Start from golang base image
FROM golang:alpine

# Add Maintainer info
LABEL maintainer="Hunter Tratar"

# Setup folders
RUN mkdir /app
WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

RUN apk add --no-cache make gcc libc-dev git

#Setup hot-reload for dev stage
RUN go install -mod=mod github.com/githubnemo/CompileDaemon
RUN go get -v golang.org/x/tools/gopls
COPY Makefile .

ENTRYPOINT CompileDaemon --build="make build" --command="./bin/api -smtp-host=localhost -smtp-port=1025 -smtp-username= -smtp-password= -metrics=true"
