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

var log logger

func Register(lev int) {
	log = newLog(lev)
}

func newLog(maxLevel int) logger {
	zerologr.SetMaxV(maxLevel)
	zl := zerolog.New(os.Stderr)
	zl = zl.With().Caller().Timestamp().Logger()
	return logger{zerologr.New(&zl)}
}

type logErr interface {
	Error(err error, msg string, keysAndValues ...any)
}

type logInfo interface {
	Info(msg string, keysAndValues ...any)
}

type LogErr interface {
	V(level int) logErr
}

type LogInf interface {
	V(level int) logInfo
}

type logErrAdapter struct {
	l logr.Logger
}

func (l logErrAdapter) V(v int) logErr {
	return l.l.V(v)
}

type logInfoAdapter struct {
	l logr.Logger
}

func (l logInfoAdapter) V(v int) logInfo {
	return l.l.V(v)
}

func LogError(logFn func(z LogErr) errro.Error) errro.Error {
	return logFn(logErrAdapter{log.Logger})
}

func LogInfo(logFn func(z LogInf)) {
	logFn(logInfoAdapter{log.Logger.WithCallDepth(2)})
}

func LogRes(logFn func(z LogErr) errro.ResError) errro.ResError {
	return LogResWithDepth(2, logFn)
}

func LogResWithDepth(depth int, logFn func(z LogErr) errro.ResError) errro.ResError {
	return logFn(logErrAdapter{log.WithCallDepth(depth)})
}
