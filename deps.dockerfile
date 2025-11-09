FROM golang:alpine3.22 AS dependencies

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
RUN go mod tidy
