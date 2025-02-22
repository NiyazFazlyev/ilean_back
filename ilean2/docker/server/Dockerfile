FROM golang:latest

WORKDIR /app

COPY ./ /app

RUN go mod download

WORKDIR /app/cmd/server

RUN go build -o main .

cmd ["./main"]