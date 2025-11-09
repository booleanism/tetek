FROM golang:alpine3.22 AS builder

WORKDIR /app
COPY --from=deps /go/pkg/mod /go/pkg/mod

COPY . .
RUN go build -ldflags="-s -w" auth/cmd/worker/main.go

FROM scratch

COPY --from=builder /app/main /auth-worker.bin

CMD [ "./auth-worker.bin" ]

EXPOSE 8082
