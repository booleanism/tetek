package messaging

import (
	"context"
	"encoding/json"

	messaging "github.com/booleanism/tetek/auth/infra/messaging/rabbitmq"
	"github.com/booleanism/tetek/auth/internal/usecases/jwt"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/go-logr/logr"
	"github.com/rabbitmq/amqp091-go"
)

type AuthContr struct {
	con *amqp091.Connection
	jwt jwt.JwtRecipes
}

func NewAuth(con *amqp091.Connection, jwtRecipe jwt.JwtRecipes) *AuthContr {
	return &AuthContr{con, jwtRecipe}
}

func (c *AuthContr) WorkerAuthListener(ctx context.Context) (*amqp091.Channel, error) {
	ctx, log := loggr.GetLogger(ctx, "worker")
	ch, err := c.con.Channel()
	if err != nil {
		log.Error(err, "failed to create channel")
		return nil, err
	}

	err = messaging.AuthSetup(ch)
	if err != nil {
		log.Error(err, "failed to setup auth topic")
		return nil, err
	}

	mgs, err := ch.Consume(messaging.AuthTaskQueue, "", false, false, false, false, nil)
	if err != nil {
		log.Error(err, "failed to consume auth task")
		return nil, err
	}

	go func() {
		for d := range mgs {
			ctx = context.WithValue(ctx, keystore.RequestID{}, d.CorrelationId)
			if d.ContentType != "text/json" {
				log.V(1).Info("unexpected ContentType", "ContentType", d.ContentType)
				helper.Nack(log, d, "unexpected ContentType", "requestID", d.CorrelationId)
				continue
			}

			task := messaging.AuthTask{}
			err := json.Unmarshal(d.Body, &task)
			if err != nil {
				log.V(1).Info("failed to marshal auth task", "requestID", d.CorrelationId, "error", err, "body", d.Body)
				res, _ := json.Marshal(messaging.AuthResult{Code: errro.ErrAuthParseFail})
				authResultPublisher(log, task, ch, d, res, err)
				continue
			}

			claim := &jwt.JwtClaims{}
			err = c.jwt.Verify(ctx, task.Jwt, &claim)
			if err != nil {
				res, _ := json.Marshal(messaging.AuthResult{Code: errro.ErrAuthJWTVerifyFail, AuthTask: task})
				authResultPublisher(log, task, ch, d, res, err)
				continue
			}

			res, _ := json.Marshal(messaging.AuthResult{Code: errro.Success, AuthTask: task, Claims: *claim})
			authResultPublisher(log, task, ch, d, res, err)
			continue
		}
	}()

	return ch, nil
}

func authResultPublisher(log logr.Logger, task messaging.AuthTask, ch *amqp091.Channel, d amqp091.Delivery, res []byte, e error) {
	log = log.WithName("auth-result-publisher")
	if err := ch.Publish(messaging.AuthExchange, messaging.AuthResRk, false, false, amqp091.Publishing{
		CorrelationId: d.CorrelationId,
		Body:          res,
		ContentType:   "text/json",
	}); err != nil {
		log.Error(err, "failed to publish auth result", "result", res, "exchange", messaging.AuthExchange, "routing-key", messaging.AuthResRk)
		helper.Nack(log, d, "unable to publish auth result", "task", task)
		return
	}

	if e != nil {
		helper.Nack(log, d, e.Error(), "task", task)
		return
	}

	if err := d.Ack(false); err != nil {
		log.Error(err, "failed to ack auth task", "task", task)
		return
	}
}
