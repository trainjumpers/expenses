package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Info, Error, Debug, Warn, Fatal func(args ...interface{})
var Infof, Errorf, Debugf, Warnf, Fatalf func(template string, args ...interface{})

func init() {
	env := os.Getenv("ENV")
	var level zap.AtomicLevel
	if env == "" {
		env = "development"
	}
	if env == "development" {
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	logger, err := zap.Config{
		Level:       level,
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "console",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalColorLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}.Build()
	if err != nil {
		panic(err)
	}
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
