package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Info, Error, Debug, Warn, Fatal func(args ...interface{})
var Infof, Errorf, Debugf, Warnf, Fatalf func(template string, args ...interface{})

func init() {
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, _ := config.Build()
	sugar := logger.Sugar()

	Info = sugar.Info
	Error = sugar.Error
	Debug = sugar.Debug
	Warn = sugar.Warn
	Fatal = sugar.Fatal

	Infof = sugar.Infof
	Errorf = sugar.Errorf
	Debugf = sugar.Debugf
	Warnf = sugar.Warnf
	Fatalf = sugar.Fatalf
}