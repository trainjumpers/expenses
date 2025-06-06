package logger

import (
	"expenses/internal/config"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Info, Error, Debug, Warn, Fatal func(args ...interface{})
var Infof, Errorf, Debugf, Warnf, Fatalf func(template string, args ...interface{})

func getLoggingLevel() zap.AtomicLevel {
	switch strings.ToLower(os.Getenv("LOGGING_LEVEL")) {
	case "debug":
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		return zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "panic":
		return zap.NewAtomicLevelAt(zap.PanicLevel)
	case "fatal":
		return zap.NewAtomicLevelAt(zap.FatalLevel)
	default:
		switch config.GetEnvironment() {
		case config.EnvironmentProd:
			return zap.NewAtomicLevelAt(zap.InfoLevel)
		case config.EnvironmentDev:
			return zap.NewAtomicLevelAt(zap.DebugLevel)
		case config.EnvironmentTest:
			return zap.NewAtomicLevelAt(zap.PanicLevel)
		default:
			return zap.NewAtomicLevelAt(zap.InfoLevel)
		}
	}
}

func getSamplingConfig() *zap.SamplingConfig {
	if config.GetEnvironment() == config.EnvironmentProd {
		return &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		}
	}
	return nil
}

func init() {
	level := getLoggingLevel()
	sampling := getSamplingConfig()

	logger, err := zap.Config{
		Level:       level,
		Development: config.GetEnvironment() != config.EnvironmentProd,
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
