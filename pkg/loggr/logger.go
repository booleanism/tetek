package loggr

import (
	"os"

	"github.com/booleanism/tetek/pkg/errro"
	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
)

type logger struct {
	zerologr.Logger
}

var Log *logger

func Register(lev int) {
	Log = newLog(lev)
}

func newLog(maxLevel int) *logger {
	zerologr.SetMaxV(maxLevel)
	zl := zerolog.New(os.Stderr)
	zl = zl.With().Caller().Timestamp().Logger()
	return &logger{zerologr.New(&zl)}
}

func (l logger) Error(v int, log func(z logr.LogSink) errro.Error) errro.Error {
	return log(Log.WithCallDepth(1).V(v).GetSink())
}
