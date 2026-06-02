package logger

import (
	"github.com/DoMinhHHung/Rental/internal/domain/port"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	z *zap.SugaredLogger
}

func New(env string) *ZapLogger {
	var cfg zap.Config
	if env == "production" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	logger, _ := cfg.Build()
	return &ZapLogger{z: logger.Sugar()}
}

func (l *ZapLogger) Info(msg string, fields ...any) {
	l.z.Infow(msg, fields...)
}

func (l *ZapLogger) Error(msg string, fields ...any) {
	l.z.Errorw(msg, fields...)
}

func (l *ZapLogger) Warn(msg string, fields ...any) {
	l.z.Warnw(msg, fields...)
}

func (l *ZapLogger) Debug(msg string, fields ...any) {
	l.z.Debugw(msg, fields...)
}

func (l *ZapLogger) With(fields ...any) port.Logger {
	return &ZapLogger{z: l.z.With(fields...)}
}
