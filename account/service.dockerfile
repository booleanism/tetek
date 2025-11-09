FROM golang:alpine3.22 AS builder

WORKDIR /app
COPY --from=deps /go/pkg/mod /go/pkg/mod

COPY . .
RUN go build -ldflags="-s -w" account/cmd/http/main.go

FROM scratch

COPY --from=builder /app/main /account.bin
COPY --from=builder /app/docs /docs

CMD [ "./account.bin" ]

EXPOSE 8082
