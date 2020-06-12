package utils

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger - used to make log faster
var Logger *zap.Logger

// Sugar - used to make log easily
var Sugar *zap.SugaredLogger

func init() {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	if debug, ok := os.LookupEnv("DEBUG"); ok && debug == "true" {
		config.Development = true
		config.Level.SetLevel(zap.DebugLevel)
	} else {
		config.Development = false
		config.Level.SetLevel(zap.InfoLevel)
	}
	Logger, _ = config.Build()
	Sugar = Logger.Sugar()
}
