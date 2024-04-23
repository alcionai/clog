package clog

import (
	"context"
	"maps"
	"os"
	"path/filepath"
	"sync"

	"github.com/alcionai/clues"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Default location for writing log files.
var userLogsDir = filepath.Join(os.Getenv("HOME"), "Library", "Logs")

// Yes, we primarily hijack zap for our logging needs here.
// This package isn't about writing a logger, it's about adding
// an opinionated shell around the zap logger.
var (
	loggerton *zap.SugaredLogger
	singleMu  sync.Mutex
)

// ---------------------------------------------------------------------------
// constructors
// ---------------------------------------------------------------------------

func genLogger(set Settings) *zap.SugaredLogger {
	// when testing, ensure debug logging matches the test.v setting
	for _, arg := range os.Args {
		if arg == `--test.v=true` {
			set.Level = LLDebug
		}
	}

	var (
		// this will be the backbone logger for the clogs
		// TODO: would be nice to accept a variety of loggers here, and
		// treat this all as a shim.  Oh well, gotta start somewhere.
		zlog *zap.Logger
		zcfg zap.Config
		// by default only add stacktraces to panics, else it gets too noisy.
		zopts = []zap.Option{zap.AddStacktrace(zapcore.PanicLevel)}
	)

	switch set.Format {
	// JSON means each row should appear as a single json object.
	case LFJSON:
		zcfg = setLevel(zap.NewProductionConfig(), set.Level)
		zcfg.OutputPaths = []string{set.File}
		// by default we'll use the columnar non-json format, which uses tab
		// separated values within each line, and may contain multiple json objs.
	default:
		zcfg = setLevel(zap.NewDevelopmentConfig(), set.Level)

		zcfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.00")

		// when printing to stdout/stderr, colorize things!
		if set.File == Stderr || set.File == Stdout {
			zcfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
	}

	zcfg.OutputPaths = []string{set.File}

	zlog, err := zcfg.Build(zopts...)
	if err != nil {
		zlog = zapcoreFallback(set)
	}

	// TODO: wrap the sugar logger to be a sugar... clogger...
	return zlog.Sugar()
}

// set up a logger core to use as a fallback in case the config doesn't work.
// we shouldn't ever need this, but it's nice to know there's a fallback in
// case configuration gets buggery, because everyone still wants their logs.
func zapcoreFallback(set Settings) *zap.Logger {
	levelFilter := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		switch set.Level {
		case LLInfo:
			return lvl >= zapcore.InfoLevel
		case LLError:
			return lvl >= zapcore.ErrorLevel
		case LLDisabled:
			return false
		default:
			// default to debug
			return true
		}
	})

	// build out the zapcore fallback
	var (
		out            = zapcore.Lock(os.Stderr)
		consoleEncoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		core           = zapcore.NewTee(zapcore.NewCore(consoleEncoder, out, levelFilter))
	)

	return zap.New(core)
}

// converts a given logLevel into the zapcore level enum.
func setLevel(cfg zap.Config, level logLevel) zap.Config {
	switch level {
	case LLInfo:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case LLError:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	case LLDisabled:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.FatalLevel)
	}

	return cfg
}

// singleton is the constructor and getter in one. Since we manage a global
// singleton for each instance, we only ever want one alive at any given time.
func singleton(set Settings) *zap.SugaredLogger {
	singleMu.Lock()
	defer singleMu.Unlock()

	if loggerton != nil {
		return loggerton
	}

	set = set.EnsureDefaults()
	setCluesSecretsHash(set.PIIHandling)

	loggerton = genLogger(set)

	return loggerton
}

// ------------------------------------------------------------------------------------------------
// context management
// ------------------------------------------------------------------------------------------------

type loggingKey string

const ctxKey loggingKey = "clog_logger"

// Init embeds a logger within the context for later retrieval.
// It is a suggested, but not necessary, initialization step.
func Init(ctx context.Context, set Settings) (context.Context, *zap.SugaredLogger) {
	zsl := singleton(set)
	zsl.Debugw("seeding logger", "logger_settings", set)

	return embedIntoCtx(ctx, zsl), zsl
}

// Seed allows users to embed their own zap.SugaredLogger within the context.
// It's good for inheriting a logger instance that was generated elsewhere, in case
// you add a downstream package that wants to clog the code.
func Seed(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	return embedIntoCtx(ctx, logger)
}

// CtxOrSeed attempts to retrieve the logger from the ctx.  If not found, it
// generates a clogger with the given logger and settings, then it to the context.
func CtxOrSeed(
	ctx context.Context,
	logger *zap.SugaredLogger,
	set Settings,
) (context.Context, *zap.SugaredLogger) {
	l := ctx.Value(ctxKey)
	if l == nil {
		zsl := singleton(set)
		return embedIntoCtx(ctx, zsl), zsl
	}

	return ctx, l.(*zap.SugaredLogger)
}

// embedIntoCtx allows users to embed their own zap.SugaredLogger within the
// context and with the given logger settings.
func embedIntoCtx(
	ctx context.Context,
	logger *zap.SugaredLogger,
) context.Context {
	if logger == nil {
		return ctx
	}

	return context.WithValue(ctx, ctxKey, logger)
}

// fromCtx pulls the clogger out of the context.  If no logger exists in the
// ctx, it returns the global singleton.
func fromCtx(ctx context.Context) *zap.SugaredLogger {
	l := ctx.Value(ctxKey)
	// if l is still nil, we need to grab the global singleton or construct a singleton.
	if l == nil {
		l = singleton(Settings{}.EnsureDefaults())
	}

	return l.(*zap.SugaredLogger)
}

// Ctx retrieves the logger embedded in the context.
// It also extracts any clues from the ctx and adds all k:v pairs to that log instance.
// TODO: Defer the ctx extraction until the time of log.
func Ctx(ctx context.Context) *zap.SugaredLogger {
	return fromCtx(ctx).With(clues.In(ctx).Slice()...)
}

// CtxErr retrieves the logger embedded in the context
// and packs all of the clues data from both the context and the error into it.
// TODO: Defer the ctx extraction until the time of log.
func CtxErr(ctx context.Context, err error) *zap.SugaredLogger {
	ctxVals := clues.In(ctx).Map()
	errVals := clues.InErr(err).Map()

	// error values should override context values.
	maps.Copy(ctxVals, errVals)

	zsl := fromCtx(ctx).
		With("error", err).
		With("error_labels", clues.Labels(err))

	for k, v := range ctxVals {
		zsl = zsl.With(k, v)
	}

	return zsl
}

// Flush writes out all buffered logs.
// Probably good to do before shutting down whatever instance is using the singleton.
func Flush(ctx context.Context) {
	_ = Ctx(ctx).Sync()
}

// ------------------------------------------------------------------------------------------------
// log wrapper for downstream api compliance
// ------------------------------------------------------------------------------------------------

type wrapper struct {
	zap.SugaredLogger
	forceDebugLogLevel bool
}

func (w *wrapper) process(opts ...option) {
	for _, opt := range opts {
		opt(w)
	}
}

type option func(*wrapper)

// ForceDebugLogLevel reduces all logs emitted in the wrapper to
// debug level, independent of their original log level.  Useful
// for silencing noisy dependency packages without losing the info
// altogether.
func ForceDebugLogLevel() option {
	return func(w *wrapper) {
		w.forceDebugLogLevel = true
	}
}

// Wrap returns the logger in the package with an extended api used for
// dependency package interface compliance.
func WrapCtx(ctx context.Context, opts ...option) *wrapper {
	return Wrap(Ctx(ctx), opts...)
}

// Wrap returns the sugaredLogger with an extended api used for
// dependency package interface compliance.
func Wrap(zsl *zap.SugaredLogger, opts ...option) *wrapper {
	w := &wrapper{SugaredLogger: *zsl}
	w.process(opts...)

	return w
}

func (w *wrapper) Logf(tmpl string, args ...any) {
	if w.forceDebugLogLevel {
		w.SugaredLogger.Debugf(tmpl, args...)
		return
	}

	w.SugaredLogger.Infof(tmpl, args...)
}

func (w *wrapper) Errorf(tmpl string, args ...any) {
	if w.forceDebugLogLevel {
		w.SugaredLogger.Debugf(tmpl, args...)
		return
	}

	w.SugaredLogger.Errorf(tmpl, args...)
}

// ------------------------------------------------------------------------------------------------
// io.writer that writes values to the logger
// ------------------------------------------------------------------------------------------------

// Writer is a wrapper that turns the logger embedded in
// the given ctx into an io.Writer.  All logs are currently
// info-level.
type Writer struct {
	Ctx context.Context
}

func (w Writer) Write(p []byte) (int, error) {
	Ctx(w.Ctx).Info(string(p))
	return len(p), nil
}
