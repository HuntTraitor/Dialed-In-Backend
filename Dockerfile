# Start from golang base image
FROM golang:alpine

# Add Maintainer info
LABEL maintainer="Hunter Tratar"

# Setup folders
RUN mkdir /app
WORKDIR /app

# Download all the dependencies
RUN go get -d -v ./...

# Install the package
RUN go install -v ./...

#Setup hot-reload for dev stage
RUN go install -mod=mod github.com/githubnemo/CompileDaemon
RUN go get -v golang.org/x/tools/gopls

ENTRYPOINT CompileDaemon --build="go build -a -installsuffix cgo -o main ./cmd/api/." --command=./main