# this image is only meant to be used for local development
FROM golang:1.17-buster

MAINTAINER Joel Mun <9oelm@wearehackerone.com>

WORKDIR /go/src/app
COPY . .

RUN go mod download

RUN go get github.com/githubnemo/CompileDaemon

# watches file change to build again
ENTRYPOINT CompileDaemon --build="go build -v" --command=./request
