FROM golang:alpine3.22 AS builder

WORKDIR /app
COPY --from=deps /go/pkg/mod /go/pkg/mod

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build go build -ldflags="-s -w" account/cmd/worker/main.go

FROM scratch

COPY --from=builder /app/main /account-worker.bin

CMD [ "./account-worker.bin" ]

EXPOSE 8082
