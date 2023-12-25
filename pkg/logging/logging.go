package logging

import (
	"context"
	"io"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger emits various logs.
type Logger = *zap.SugaredLogger

// LoggerFactory retrieves a named logger for a given module.
type LoggerFactory func(module string) Logger

// Module returns an function that returns a logger for a given module when provided with a context.
func Module(module string) func(ctx context.Context) Logger {
	return func(ctx context.Context) Logger {
		if l := ctx.Value(loggerCacheKey); l != nil {
			return l.(*loggerCache).getLogger(module) //nolint:forcetypeassert
		}

		return NullLogger
	}
}

// ToWriter returns LoggerFactory that uses given writer for log output (unadorned).
func ToWriter(w io.Writer) LoggerFactory {
	return zap.New(zapcore.NewCore(
		NewStdConsoleEncoder(StdConsoleEncoderConfig{}),
		zapcore.AddSync(w), zap.DebugLevel), zap.WithClock(Clock())).Sugar().Named
}
