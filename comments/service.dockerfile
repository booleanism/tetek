FROM golang:alpine3.22 AS builder

WORKDIR /app
COPY --from=deps /go/pkg/mod /go/pkg/mod

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build go build -ldflags="-s -w" comments/cmd/http/main.go

FROM scratch

COPY --from=builder /app/main /comments.bin
COPY --from=builder /app/docs /docs

CMD [ "./comments.bin" ]

EXPOSE 8082
