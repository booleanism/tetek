package main

import (
	"os"

	"github.com/booleanism/tetek/auth/cmd/http/router"
	"github.com/booleanism/tetek/auth/internal/contract"
	"github.com/booleanism/tetek/auth/internal/jwt"
	"github.com/booleanism/tetek/auth/recipes"
	"github.com/gofiber/fiber/v3"
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
	defer mqCon.Close()

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
	defer ch.Close()

	accContr := contract.NewAccount(mqCon)
	logRec := recipes.NewLogin(accContr, jwt)

	app := fiber.New()
	api := app.Group("/api/v0")
	{
		api.Post("/", router.Login(logRec))
	}

	if err := app.Listen(":8081"); err != nil {
		panic(err)
	}
}
