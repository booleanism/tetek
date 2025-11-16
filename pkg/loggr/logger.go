package loggr

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
)

func NewLogger(name string, base *zerolog.Logger) logr.Logger {
	return zerologr.New(base).WithName(name).V(0)
}

func GetLogger(ctx context.Context, scope string) (context.Context, logr.Logger) {
	log, _ := logr.FromContext(ctx)
	// up log to use with serive name
	log = log.WithName(scope)

	ctx = logr.NewContext(ctx, log)
	return ctx, log
}
