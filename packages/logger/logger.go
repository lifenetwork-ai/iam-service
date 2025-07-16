package logger

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Level defines the severity level for logging.
type Level string

const (
	DebugLevel Level = "debug"
	InfoLevel  Level = "info"
	WarnLevel  Level = "warn"
	ErrorLevel Level = "error"
	FatalLevel Level = "fatal"
	PanicLevel Level = "panic"
)

// Logger interface defines common logging operations.
type Logger interface {
	// Level management
	SetLogLevel(level Level)
	GetLogLevel() Level

	// Basic logging methods
	Debug(message string)
	Debugf(format string, values ...any)
	Info(message string)
	Infof(format string, values ...any)
	Warn(message string)
	Warnf(format string, values ...any)
	Error(message string)
	Errorf(format string, values ...any)
	Fatal(message string)
	Fatalf(format string, values ...any)
	Panic(message string)
	Panicf(format string, values ...any)

	// Contextual logging
	WithFields(fields map[string]any) Logger
}

type zapLogger struct {
	sugaredLogger *zap.SugaredLogger
	currentLevel  zap.AtomicLevel
}

var (
	instance *zapLogger
	once     sync.Once
)

// mapLogLevel maps interfaces.Level to zapcore.Level.
func mapLogLevel(level Level) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	case PanicLevel:
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
}

// customLevelEncoder replaces "level" with "severity".
func customLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	severityMapping := map[zapcore.Level]string{
		zapcore.DebugLevel: "DEBUG",
		zapcore.InfoLevel:  "INFO",
		zapcore.WarnLevel:  "WARNING",
		zapcore.ErrorLevel: "ERROR",
		zapcore.PanicLevel: "CRITICAL",
		zapcore.FatalLevel: "ALERT",
	}
	enc.AppendString(severityMapping[level])
}

// newZapLogger initializes a new zap-based logger instance.
func newZapLogger(level Level) *zapLogger {
	atomicLevel := zap.NewAtomicLevelAt(mapLogLevel(level))

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:      "timestamp",
		LevelKey:     "severity",
		CallerKey:    "caller",
		MessageKey:   "message",
		EncodeLevel:  customLevelEncoder,
		EncodeTime:   zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05Z07:00"),
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.Lock(zapcore.AddSync(os.Stdout)),
		atomicLevel,
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return &zapLogger{
		sugaredLogger: logger.Sugar(),
		currentLevel:  atomicLevel,
	}
}

// GetLogger returns the singleton logger instance.
func GetLogger() Logger {
	once.Do(func() {
		instance = newZapLogger(InfoLevel)
	})
	return instance
}

// SetLogger replaces the singleton logger instance.
func SetLogger(customLogger Logger) {
	instance = customLogger.(*zapLogger)
}

// SetLogLevel sets the log level dynamically.
func SetLogLevel(level Level) {
	if instance != nil {
		instance.currentLevel.SetLevel(mapLogLevel(level))
	}
}

// Logger Interface Implementation
func (z *zapLogger) SetLogLevel(level Level) {
	SetLogLevel(level)
}

func (z *zapLogger) GetLogLevel() Level {
	return Level(z.currentLevel.String())
}

// Individual Log Methods
func (z *zapLogger) Debug(message string) { z.sugaredLogger.Debug(message) }

func (z *zapLogger) Debugf(format string, values ...interface{}) {
	z.sugaredLogger.Debugf(format, values...)
}
func (z *zapLogger) Info(message string) { z.sugaredLogger.Info(message) }
func (z *zapLogger) Infof(format string, values ...interface{}) {
	z.sugaredLogger.Infof(format, values...)
}
func (z *zapLogger) Warn(message string) { z.sugaredLogger.Warn(message) }
func (z *zapLogger) Warnf(format string, values ...interface{}) {
	z.sugaredLogger.Warnf(format, values...)
}
func (z *zapLogger) Error(message string) { z.sugaredLogger.Error(message) }
func (z *zapLogger) Errorf(format string, values ...interface{}) {
	z.sugaredLogger.Errorf(format, values...)
}
func (z *zapLogger) Fatal(message string) { z.sugaredLogger.Fatal(message) }
func (z *zapLogger) Fatalf(format string, values ...interface{}) {
	z.sugaredLogger.Fatalf(format, values...)
}
func (z *zapLogger) Panic(message string) { z.sugaredLogger.Panic(message) }
func (z *zapLogger) Panicf(format string, values ...interface{}) {
	z.sugaredLogger.Panicf(format, values...)
}

// WithFields creates a new logger instance with additional fields.
func (z *zapLogger) WithFields(fields map[string]interface{}) Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}

	return &zapLogger{
		sugaredLogger: z.sugaredLogger.With(args...),
		currentLevel:  z.currentLevel,
	}
}
