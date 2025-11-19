package main

import (
	"os"

	"github.com/booleanism/tetek/auth/cmd/http/router"
	"github.com/booleanism/tetek/auth/internal/contract"
	"github.com/booleanism/tetek/auth/internal/jwt"
	"github.com/booleanism/tetek/auth/recipes"
	"github.com/booleanism/tetek/pkg/contracts"
	"github.com/booleanism/tetek/pkg/helper/http/middlewares"
	"github.com/go-logr/zerologr"
	"github.com/gofiber/fiber/v3"
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
	defer func() {
		if err := mqCon.Close(); err != nil {
			panic(err)
		}
	}()

	jwtSecret := os.Getenv("AUTH_JWT_SECRET")
	if jwtSecret == "" {
		panic("jwt secret empty")
	}

	jwt := jwt.NewJwt([]byte(jwtSecret))
	auth := contract.NewAuth(mqCon, jwt)
	ch, err := auth.WorkerAuthListener()
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := ch.Close(); err != nil {
			panic(err)
		}
	}()

	accContr := contracts.SubscribeAccount(mqCon, "auth")
	logRec := recipes.NewLogin(accContr, jwt)

	app := fiber.New()
	api := app.Group("/api/v0")
	{
		api.Use(middlewares.GenerateRequestID)
		api.Use(middlewares.Logger(ServiceName, &zl))
		api.Post("/", router.Login(logRec))
	}

	if err := app.Listen(":8081"); err != nil {
		panic(err)
	}
}
