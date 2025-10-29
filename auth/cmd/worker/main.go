package main

import (
	"os"

	"github.com/booleanism/tetek/auth/internal/contract"
	"github.com/booleanism/tetek/auth/internal/jwt"
	"github.com/rabbitmq/amqp091-go"
)

func main() {
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
	ch, err := authCtr.WorkerAuthListener()
	if err != nil {
		panic(err)
	}
	defer ch.Close()
}
