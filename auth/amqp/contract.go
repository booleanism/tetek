package amqp

import (
	"github.com/booleanism/tetek/auth/internal/jwt"
	"github.com/rabbitmq/amqp091-go"
)

const (
	AUTH_EXCHANGE       = "x_auth"
	AUTH_TASK_QUEUE     = "task_auth"
	AUTH_TASK_RK        = "task.auth.*"
	AUTH_RES_QUEUE      = "res_auth"
	AUTH_RES_RK         = "res.auth.*"
	DLX_AUTH_EXCHANGE   = "dlx_auth"
	DLQ_AUTH_TASK_QUEUE = "dlq_task_auth"
)

type AuthTask struct {
	Jwt string `json:"jwt"`
}

type AuthResult struct {
	Code   int           `json:"code"`
	Claims jwt.JwtClaims `json:"claim"`
	AuthTask
}

func AuthSetup(ch *amqp091.Channel) error {
	err := ch.ExchangeDeclare(AUTH_EXCHANGE, "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare(DLX_AUTH_EXCHANGE, "fanout", true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(DLQ_AUTH_TASK_QUEUE, true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.QueueBind(DLQ_AUTH_TASK_QUEUE, "", DLX_AUTH_EXCHANGE, false, nil)
	if err != nil {
		return err
	}

	qTask, err := ch.QueueDeclare(AUTH_TASK_QUEUE, true, false, false, false, amqp091.Table{
		"x-dead-letter-exchange": DLX_AUTH_EXCHANGE,
		"x-message-ttl":          int32(5000),
	})
	if err != nil {
		return err
	}

	err = ch.QueueBind(qTask.Name, AUTH_TASK_RK, AUTH_EXCHANGE, false, nil)
	if err != nil {
		return err
	}

	return nil
}
