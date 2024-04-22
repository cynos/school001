# Start from golang base image
FROM golang:1.18.3-buster

LABEL maintainer="Setia Budi"

RUN go install github.com/githubnemo/CompileDaemon@latest

RUN mkdir /app

WORKDIR /app

COPY . .

RUN go get -d -v ./...

RUN go mod tidy

RUN go mod vendor

ENTRYPOINT CompileDaemon --build="go build -a -installsuffix cgo -o main ." --command="./main --debug=true --basepath=/home/developer/budi/dist/ --config=/config.json"