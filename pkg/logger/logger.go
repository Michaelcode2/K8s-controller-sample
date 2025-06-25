package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger wraps zerolog.Logger for easier usage
type Logger struct {
	logger zerolog.Logger
}

// New creates a new logger instance based on environment
func New() *Logger {
	// Set global log level based on environment
	env := os.Getenv("ENV")
	if env == "" {
		env = "dev" // default to dev environment
	}

	var level zerolog.Level
	switch env {
	case "prod", "production":
		level = zerolog.InfoLevel
		// In production, use JSON format for better parsing
		zerolog.TimeFieldFormat = time.RFC3339
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	default:
		// In development, use pretty console output
		level = zerolog.DebugLevel
		zerolog.TimeFieldFormat = "15:04:05"
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
			FormatLevel: func(i interface{}) string {
				if ll, ok := i.(string); ok {
					switch ll {
					case "debug":
						return "🔍"
					case "info":
						return "ℹ️"
					case "warn":
						return "⚠️"
					case "error":
						return "❌"
					case "fatal":
						return "💀"
					case "panic":
						return "🚨"
					}
				}
				return "?"
			},
		})
	}

	zerolog.SetGlobalLevel(level)

	// Add some default fields
	logger := log.With().
		Str("service", "k8s-controller").
		Str("environment", env).
		Logger()

	return &Logger{logger: logger}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	event := l.logger.Debug()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	event := l.logger.Info()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	event := l.logger.Warn()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Error logs an error message
func (l *Logger) Error(msg string, err error, fields map[string]interface{}) {
	event := l.logger.Error()
	if err != nil {
		event = event.Err(err)
	}
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, err error, fields map[string]interface{}) {
	event := l.logger.Fatal()
	if err != nil {
		event = event.Err(err)
	}
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// WithNamespace returns a logger with namespace field
func (l *Logger) WithNamespace(namespace string) *Logger {
	return &Logger{
		logger: l.logger.With().Str("namespace", namespace).Logger(),
	}
}

// WithDeployment returns a logger with deployment field
func (l *Logger) WithDeployment(deploymentName string) *Logger {
	return &Logger{
		logger: l.logger.With().Str("deployment", deploymentName).Logger(),
	}
}

// GetZerologLogger returns the underlying zerolog.Logger for advanced usage
func (l *Logger) GetZerologLogger() zerolog.Logger {
	return l.logger
}
