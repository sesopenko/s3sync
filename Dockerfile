# syntax=docker/dockerfile:1

FROM golang:1.21-bullseye

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /s3sync

CMD ["/s3sync"]