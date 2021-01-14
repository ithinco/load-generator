package helper

import "go.uber.org/zap"

var Logger = newLogger()

func newLogger() *zap.Logger {
	logger, _ := zap.NewProduction()

	return logger
}
