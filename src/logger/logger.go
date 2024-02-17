package logger

import "go.uber.org/zap"

var Sugar *zap.SugaredLogger

func init() {
	// Initialize the logger
	logger, _ := zap.NewProduction() // Consider handling the error in real applications
	Sugar = logger.Sugar()
}
