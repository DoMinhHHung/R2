package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	z *zap.SugaredLogger
}

func New(env string) *Logger {
	var cfg zap.Config

	if env == "production" {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "ts"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	l, _ := cfg.Build(zap.AddCallerSkip(1))
	return &Logger{z: l.Sugar()}
}

func (l *Logger) Info(msg string, fields ...any)  { l.z.Infow(msg, fields...) }
func (l *Logger) Error(msg string, fields ...any) { l.z.Errorw(msg, fields...) }
func (l *Logger) Warn(msg string, fields ...any)  { l.z.Warnw(msg, fields...) }
func (l *Logger) Debug(msg string, fields ...any) { l.z.Debugw(msg, fields...) }

func (l *Logger) With(fields ...any) *Logger {
	return &Logger{z: l.z.With(fields...)}
}
