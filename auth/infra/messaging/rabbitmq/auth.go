package messaging

import (
	"github.com/booleanism/tetek/auth/internal/usecases/jwt"
	"github.com/rabbitmq/amqp091-go"
)

const (
	AuthExchange     = "x_auth"
	AuthTaskQueue    = "task_auth"
	AuthTaskRk       = "task.auth.*"
	AuthResQueue     = "res_auth"
	AuthResRk        = "res.auth.*"
	DlxAuthExchange  = "dlx_auth"
	DlqAuthTaskQueue = "dlq_task_auth"
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
	err := ch.ExchangeDeclare(AuthExchange, "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare(DlxAuthExchange, "fanout", true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(DlqAuthTaskQueue, true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.QueueBind(DlqAuthTaskQueue, "", DlxAuthExchange, false, nil)
	if err != nil {
		return err
	}

	qTask, err := ch.QueueDeclare(AuthTaskQueue, true, false, false, false, amqp091.Table{
		"x-dead-letter-exchange": DlxAuthExchange,
		"x-message-ttl":          int32(5000),
	})
	if err != nil {
		return err
	}

	err = ch.QueueBind(qTask.Name, AuthTaskRk, AuthExchange, false, nil)
	if err != nil {
		return err
	}

	return nil
}
