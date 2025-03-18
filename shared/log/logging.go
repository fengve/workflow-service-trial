package log

import (
	"context"
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"go.temporal.io/sdk/workflow"
)

// Common context keys in the request's user context.
const (
	CTX_KEY_LOGGER  = "logger"
	CTX_KEY_ORGID   = "orgId"
	CTX_KEY_TRACEID = "traceId"
)

type Logger interface {
	Debug(msg string, keyvals ...interface{})
	Info(msg string, keyvals ...interface{})
	Warn(msg string, keyvals ...interface{})
	Error(msg string, keyvals ...interface{})

	With(key string, value interface{}) Logger
}

type DefaultLogger struct {
	logger log.Logger
}

func (l DefaultLogger) Debug(msg string, keyvals ...interface{}) {
	_ = level.Debug(l.logger).Log(append([]interface{}{"msg", msg}, keyvals...)...)
}
func (l DefaultLogger) Info(msg string, keyvals ...interface{}) {
	_ = level.Info(l.logger).Log(append([]interface{}{"msg", msg}, keyvals...)...)
}
func (l DefaultLogger) Warn(msg string, keyvals ...interface{}) {
	_ = level.Warn(l.logger).Log(append([]interface{}{"msg", msg}, keyvals...)...)
}
func (l DefaultLogger) Error(msg string, keyvals ...interface{}) {
	_ = level.Error(l.logger).Log(append([]interface{}{"msg", msg}, keyvals...)...)
}
func (l DefaultLogger) With(key string, value interface{}) Logger {
	return DefaultLogger{logger: log.With(l.logger, key, value)}
}

// Interface to cover both context.Context and workflow.Context
type ContextWithValue interface {
	Value(key interface{}) interface{}
}

func GetLogger(ctx ContextWithValue) Logger {
	if ctx == nil {
		return createLogger(context.Background())
	}
	if logger, ok := ctx.Value(CTX_KEY_LOGGER).(Logger); ok {
		return logger
	}
	return createLogger(ctx)
}

func SetLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, CTX_KEY_LOGGER, logger)
}

func SetWorkflowLogger(ctx workflow.Context, logger Logger) workflow.Context {
	return workflow.WithValue(ctx, CTX_KEY_LOGGER, logger)
}

func createLogger(ctx ContextWithValue) Logger {
	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
	logger = level.NewFilter(logger, level.AllowInfo()) // Determine from env var.

	// TraceId
	if val, ok := ctx.Value(CTX_KEY_TRACEID).(string); ok {
		logger = log.With(logger, CTX_KEY_TRACEID, val)
	}

	// OrgId
	if val, ok := ctx.Value(CTX_KEY_ORGID).(string); ok {
		logger = log.With(logger, CTX_KEY_ORGID, val)
	}

	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.Caller(4))

	return DefaultLogger{logger: logger}
}
