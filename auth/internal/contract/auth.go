package contract

import (
	"encoding/json"

	"github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/auth/internal/jwt"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/rabbitmq/amqp091-go"
)

type AuthContr struct {
	con *amqp091.Connection
	jwt jwt.JwtRecipes
}

func NewAuth(con *amqp091.Connection, jwtRecipe jwt.JwtRecipes) *AuthContr {
	return &AuthContr{con, jwtRecipe}
}

func (c *AuthContr) WorkerAuthListener() (*amqp091.Channel, error) {
	ch, err := c.con.Channel()
	if err != nil {
		return nil, err
	}

	err = amqp.AuthSetup(ch)
	if err != nil {
		return nil, err
	}

	mgs, err := ch.Consume(amqp.AUTH_TASK_QUEUE, "", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	go func() {
		for d := range mgs {
			if d.ContentType != "text/json" {
				continue
			}

			task := amqp.AuthTask{}
			err := json.Unmarshal(d.Body, &task)
			if err != nil {
				res, _ := json.Marshal(amqp.AuthResult{Code: errro.EAUTH_PARSE_FAIL})
				ch.Publish(amqp.AUTH_EXCHANGE, amqp.AUTH_RES_RK, false, false, amqp091.Publishing{
					CorrelationId: d.CorrelationId,
					Body:          res,
					ContentType:   "text/json",
				})
				d.Nack(false, false)
				continue
			}

			claim, err := c.jwt.Verify(task.Jwt)
			if err != nil {
				res, _ := json.Marshal(amqp.AuthResult{Code: errro.EAUTH_JWT_VERIFY_FAIL, AuthTask: task})
				ch.Publish(amqp.AUTH_EXCHANGE, amqp.AUTH_RES_RK, false, false, amqp091.Publishing{
					CorrelationId: d.CorrelationId,
					Body:          res,
					ContentType:   "text/json",
				})
				d.Nack(false, false)
				continue
			}

			res, _ := json.Marshal(amqp.AuthResult{Code: errro.SUCCESS, AuthTask: task, Claims: *claim})
			ch.Publish(amqp.AUTH_EXCHANGE, amqp.AUTH_RES_RK, false, false, amqp091.Publishing{
				CorrelationId: d.CorrelationId,
				Body:          res,
				ContentType:   "text/json",
			})
			d.Ack(false)
		}
	}()

	return ch, nil
}
