package logger

import (
	"io"
	"os"
	"time"

	zap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	logger *zap.Logger
	level  Level
}

type Level = zapcore.Level

const (
	InfoLevel Level = zap.InfoLevel // 0, default level
)

type Field = zap.Field

func (logger *Logger) Debug(msg string, fields ...Field) {
	logger.logger.Debug(msg, fields...)
}

func (logger *Logger) Info(msg string, fields ...Field) {
	logger.logger.Info(msg, fields...)
}

func (logger *Logger) Warn(msg string, fields ...Field) {
	logger.logger.Warn(msg, fields...)
}

func (logger *Logger) Error(msg string, fields ...Field) {
	logger.logger.Error(msg, fields...)
}

func (logger *Logger) DPanic(msg string, fields ...Field) {
	logger.logger.DPanic(msg, fields...)
}

func (logger *Logger) Panic(msg string, fields ...Field) {
	logger.logger.Panic(msg, fields...)
}

func (logger *Logger) Fatal(msg string, fields ...Field) {
	logger.logger.Fatal(msg, fields...)
}

// function variables for all field types
// in github.com/uber-go/zap/field.go
var (
	Skip        = zap.Skip
	Binary      = zap.Binary
	Bool        = zap.Bool
	Boolp       = zap.Boolp
	ByteString  = zap.ByteString
	Complex128  = zap.Complex128
	Complex128p = zap.Complex128p
	Complex64   = zap.Complex64
	Complex64p  = zap.Complex64p
	Float64     = zap.Float64
	Float64p    = zap.Float64p
	Float32     = zap.Float32
	Float32p    = zap.Float32p
	Int         = zap.Int
	Intp        = zap.Intp
	Int64       = zap.Int64
	Int64p      = zap.Int64p
	Int32       = zap.Int32
	Int32p      = zap.Int32p
	Int16       = zap.Int16
	Int16p      = zap.Int16p
	Int8        = zap.Int8
	Int8p       = zap.Int8p
	String      = zap.String
	Stringp     = zap.Stringp
	Uint        = zap.Uint
	Uintp       = zap.Uintp
	Uint64      = zap.Uint64
	Uint64p     = zap.Uint64p
	Uint32      = zap.Uint32
	Uint32p     = zap.Uint32p
	Uint16      = zap.Uint16
	Uint16p     = zap.Uint16p
	Uint8       = zap.Uint8
	Uint8p      = zap.Uint8p
	Uintptr     = zap.Uintptr
	Uintptrp    = zap.Uintptrp
	Reflect     = zap.Reflect
	Namespace   = zap.Namespace
	Stringer    = zap.Stringer
	Time        = zap.Time
	Timep       = zap.Timep
	Stack       = zap.Stack
	StackSkip   = zap.StackSkip
	Duration    = zap.Duration
	Durationp   = zap.Durationp
	Any         = zap.Any
	NamedError  = zap.NamedError

	Info   = std.Info
	Warn   = std.Warn
	Error  = std.Error
	DPanic = std.DPanic
	Panic  = std.Panic
	Fatal  = std.Fatal
	Debug  = std.Debug
)

func ResetDefault(logger *Logger) {
	std = logger
	Info = std.Info
	Warn = std.Warn
	Error = std.Error
	DPanic = std.DPanic
	Panic = std.Panic
	Fatal = std.Fatal
	Debug = std.Debug
}

var std = New(os.Stderr, InfoLevel, WithCaller(true))

func Default() *Logger {
	return std
}

type Option = zap.Option

var (
	WithCaller    = zap.WithCaller
	AddStacktrace = zap.AddStacktrace
)

type RotateOptions struct {
	MaxSize    int
	MaxAge     int
	MaxBackups int
	Compress   bool
}

type LevelEnablerFunc func(lvl Level) bool

type TeeOption struct {
	Filename string
	Ropt     RotateOptions
	Lef      LevelEnablerFunc
}

func NewTeeWithRotate(tops []TeeOption, opts ...Option) *Logger {
	var cores []zapcore.Core

	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02T15:04:05.000Z0700"))
	}

	for _, top := range tops {
		top := top

		lv := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return top.Lef(Level(lvl))
		})

		writer := zapcore.AddSync(&lumberjack.Logger{
			Filename:   top.Filename,
			MaxSize:    top.Ropt.MaxSize,
			MaxBackups: top.Ropt.MaxBackups,
			MaxAge:     top.Ropt.MaxAge,
			Compress:   top.Ropt.Compress,
		})

		// Print in file.
		coreFile := zapcore.NewCore(
			zapcore.NewJSONEncoder(cfg.EncoderConfig),
			zapcore.AddSync(writer),
			lv,
		)

		// Print in console.
		coreConsole := zapcore.NewCore(
			zapcore.NewConsoleEncoder(cfg.EncoderConfig),
			zapcore.AddSync(os.Stdout),
			lv,
		)

		cores = append(cores, coreFile, coreConsole)
	}

	return &Logger{
		logger: zap.New(zapcore.NewTee(cores...), opts...),
	}
}

// New create a new logger (not support log rotating).
func New(writer io.Writer, level Level, opts ...Option) *Logger {
	if writer == nil {
		panic("the writer is nil")
	}

	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02T15:04:05.000Z0700"))
	}

	lv := zapcore.Level(level)

	// Print in file.
	coreFile := zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg.EncoderConfig),
		zapcore.AddSync(writer),
		lv,
	)

	// Print in console.
	coreConsole := zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg.EncoderConfig),
		zapcore.AddSync(os.Stdout),
		lv,
	)

	var cores []zapcore.Core
	cores = append(cores, coreFile, coreConsole)

	return &Logger{
		logger: zap.New(zapcore.NewTee(cores...), opts...),
		level:  level,
	}
}

func (logger *Logger) Sync() error {
	return logger.logger.Sync()
}

func Sync() error {
	if std != nil {
		return std.Sync()
	}

	return nil
}
