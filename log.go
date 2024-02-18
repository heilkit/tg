package tg

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

type Logger interface {
	OnHandle(end string, ctx Context, duration time.Duration)

	OnError(err error, ctx Context)

	OnRaw(method string, payload []byte, response []byte, err error, duration time.Duration)
}

func LoggerSlog(logger ...*slog.Logger) Logger {
	if len(logger) == 0 {
		return &loggerSlog{slog.Default()}
	}
	return &loggerSlog{logger[0]}
}

type loggerSlog struct {
	logger *slog.Logger
}

func (logger loggerSlog) OnHandle(end string, ctx Context, duration time.Duration) {
	args := []any{"time", duration.String(), "endpoint", end}
	if buff, err := json.Marshal(ctx.Chat()); err == nil {
		args = append(args, "chat", string(buff))
	}
	if buff, err := json.Marshal(ctx.Message()); err == nil {
		args = append(args, "message", string(buff))
	}
	logger.logger.Info("handle", args...)
}

func (logger loggerSlog) OnRaw(method string, payload []byte, response []byte, err error, duration time.Duration) {
	logger.logger.Debug("raw",
		"time", duration.String(),
		"method", method,
		"payload", string(payload),
		"response", string(response),
		"error", fmt.Sprintf("%v", err),
	)
}

func (logger loggerSlog) OnError(err error, ctx Context) {
	args := []any{"err", fmt.Sprintf("%v", err)}
	if buff, err := json.Marshal(ctx.Chat()); err == nil {
		args = append(args, "chat", string(buff))
	}
	if buff, err := json.Marshal(ctx.Message()); err == nil {
		args = append(args, "message", string(buff))
	}
	logger.logger.Info("error", args...)
}
