package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger constructs a Sugared Logger that writes to stdout and
// a provided log file. Provides human readable timestamps.
func NewLogger(label string, file string, term bool) *zap.SugaredLogger {
	config := zap.NewProductionConfig()
	config.Encoding = "json"
	// set output paths. If term is true, include terminal output
	if term {
		config.OutputPaths = []string{"stdout", file}
	} else {
		config.OutputPaths = []string{file}
	}
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.DisableStacktrace = true
	config.InitialFields = map[string]interface{}{
		"LABEL": label,
	}

	logger, err := config.Build()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return logger.Sugar()
}
