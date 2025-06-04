package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Info, Error, Debug, Warn, Fatal func(args ...interface{})
var Infof, Errorf, Debugf, Warnf, Fatalf func(template string, args ...interface{})

func init() {
	env := strings.ToLower(os.Getenv("ENV"))
	var level zap.AtomicLevel
	var sampling *zap.SamplingConfig
	if env == "" {
		env = "dev"
	}
	if env == "prod" {
		level = zap.NewAtomicLevelAt(zap.InfoLevel)
		sampling = &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		}
	} else if env == "test" {
		level = zap.NewAtomicLevelAt(zap.PanicLevel)
		sampling = nil
	} else {
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
		sampling = nil
	}

	logger, err := zap.Config{
		Level:       level,
		Development: env != "prod",
		Sampling:    sampling,
		Encoding:    "console",
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
