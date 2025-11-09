FROM golang:alpine3.22 AS builder

WORKDIR /app
COPY --from=deps /go/pkg/mod /go/pkg/mod

COPY . .
RUN go build -ldflags="-s -w" feeds/cmd/http/main.go

FROM scratch

COPY --from=builder /app/main /feeds.bin
COPY --from=builder /app/docs /docs

CMD [ "./feeds.bin" ]

EXPOSE 8083
