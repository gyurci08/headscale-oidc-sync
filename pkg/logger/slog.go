package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"

	"hu.jandzsogyorgy.headscale-oidc-sync/pkg/config"
)

type Logger struct {
	logger *slog.Logger
}

var _ ILogger = (*Logger)(nil)

func NewLogger(cfg config.Config, writer io.Writer) (ILogger, error) {
	if writer == nil {
		writer = os.Stdout
	}

	logCfg := cfg.Log

	var level slog.Level
	if err := level.UnmarshalText([]byte(logCfg.Level)); err != nil {
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	}

	var handler slog.Handler
	switch strings.ToLower(logCfg.Format) {
	case "json":
		handler = slog.NewJSONHandler(writer, opts)
	case "text":
		handler = slog.NewTextHandler(writer, opts)
	case "console":
		handler = NewConsoleHandler(writer, opts)
	default:
		handler = slog.NewTextHandler(writer, opts)
	}

	loggerInstance := slog.New(handler)
	return &Logger{logger: loggerInstance}, nil
}

func (s *Logger) Debug(msg string, args ...any) {
	s.logger.Debug(msg, args...)
}

func (s *Logger) Info(msg string, args ...any) {
	s.logger.Info(msg, args...)
}

func (s *Logger) Warn(msg string, args ...any) {
	s.logger.Warn(msg, args...)
}

func (s *Logger) Error(msg string, args ...any) {
	s.logger.Error(msg, args...)
}

func (s *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	s.logger.DebugContext(ctx, msg, args...)
}

func (s *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	s.logger.InfoContext(ctx, msg, args...)
}

func (s *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	s.logger.WarnContext(ctx, msg, args...)
}

func (s *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	s.logger.ErrorContext(ctx, msg, args...)
}

func (s *Logger) With(args ...any) ILogger {
	return &Logger{
		logger: s.logger.With(args...),
	}
}
