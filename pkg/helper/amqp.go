package helper

import (
	"github.com/go-logr/logr"
	"github.com/rabbitmq/amqp091-go"
)

func Nack(log logr.Logger, d amqp091.Delivery, reason string, keysAndValues ...any) {
	if err := d.Nack(false, false); err != nil {
		log.Error(err, "failed to nack message: "+reason, keysAndValues)
	}
}
