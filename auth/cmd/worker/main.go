package main

import (
	"context"
	"os"

	"github.com/booleanism/tetek/auth/internal/contract"
	"github.com/booleanism/tetek/auth/internal/jwt"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

const (
	ServiceName = "auth"
	LogV        = 4
)

func main() {
	zl := zerolog.New(os.Stderr)
	zerologr.SetMaxV(LogV)
	mqStr := os.Getenv("BROKER_STR")
	if mqStr == "" {
		panic("amqp connection string empty")
	}

	mqCon, err := amqp091.Dial(mqStr)
	if err != nil {
		panic(err)
	}

	jwtSecret := os.Getenv("AUTH_JWT_SECRET")
	if jwtSecret == "" {
		panic("jwt secret empty")
	}

	jwt := jwt.NewJwt([]byte(jwtSecret))
	authCtr := contract.NewAuth(mqCon, jwt)

	workerCtx := context.Background()
	workerCtx = logr.NewContext(workerCtx, loggr.NewLogger(ServiceName, &zl))
	ch, err := authCtr.WorkerAuthListener(workerCtx)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := ch.Close(); err != nil {
			panic(err)
		}
	}()
	select {}
}
