package utils

import (
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
	Logger, _ = config.Build()
	Sugar = Logger.Sugar()
}
