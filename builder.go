package clog

import (
	"context"

	"github.com/alcionai/clues"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
)

// ------------------------------------------------------------------------------------------------
// builder is the primary logging handler
// most funcs that people would use in the daily drive are going
// to modfiy and/or return a builder instance.  The builder aggregates
// data passed to it until a log func is called (debug, info, or error).
// At that time it consumes all of the gathered data to send the log message.
// ------------------------------------------------------------------------------------------------

type builder struct {
	ctx      context.Context
	err      error
	zsl      *zap.SugaredLogger
	with     map[any]any
	labels   map[string]struct{}
	comments map[string]struct{}
}

func newBuilder(ctx context.Context) *builder {
	zsl := fromCtx(ctx)

	return &builder{
		ctx: ctx,
		zsl: zsl,
	}
}

type level string

var (
	lvlDebug level = "debug"
	lvlInfo  level = "info"
	lvlError level = "error"
)

// log actually delivers the log to the underlying logger with the given
func (b builder) log(l level, msg string) {
	cv := clues.In(b.ctx).Map()
	zsl := b.zsl

	if b.err != nil {
		// error values should override context values.
		maps.Copy(cv, clues.InErr(b.err).Map())

		// attach the error and its labels
		zsl = zsl.
			With("error", b.err).
			With("error_labels", clues.Labels(b.err))
	}

	// pack in all clues and error values
	for k, v := range cv {
		zsl = zsl.With(k, v)
	}

	// plus any values added using builder.With()
	for k, v := range b.with {
		zsl = zsl.With(k, v)
	}

	// finally, make sure we attach the labels and comments
	zsl = zsl.With("clog_labels", maps.Keys(b.labels))
	zsl = zsl.With("clog_comments", maps.Keys(b.comments))

	// then write everything to the logger
	switch l {
	case lvlDebug:
		zsl.Debug(msg)
	case lvlInfo:
		zsl.Info(msg)
	case lvlError:
		zsl.Error(msg)
	}
}

// Err attaches the error to the builder.
// When logged, the error will be parsed for any clues parts
// and those values will get added to the resulting log.
//
// ex: if you have some `err := clues.New("badness").With("cause", reason)`
// then this will add both of the following to the log:
// - "error": "badness"
// - "cause": reason
func (b *builder) Err(err error) *builder {
	b.err = err
	return b
}

// Label adds all of the appended labels to the error.
// Adding labels is a great way to categorize your logs into broad scale
// concepts like "configuration", "process kickoff", or "process conclusion".
// they're also a great way to set up various granularities of debugging
// like "api queries" or "fine grained item review", since you  can configure
// clog to automatically filter debug level logging to only deliver if the
// logs match one or more labels, allowing you to only emit some of the
// overwhelming number of debug logs that we all know you produce, you
// little overlogger, you.
func (b *builder) Label(ls ...string) *builder {
	for _, l := range ls {
		b.labels[l] = struct{}{}
	}

	return b
}

// Comments are available because why make your devs go all the way back to
// the code to find the comment about this log case?  Add them into the log
// itself!
func (b *builder) Comment(cmnt string) *builder {
	b.comments[cmnt] = struct{}{}
	return b
}

// With is your standard "With" func.  Add data in K:V pairs here to have them
// added to the log message metadata.  Ex: builder.With("foo", "bar") will add
// "foo": "bar" to the resulting log structure.  An uneven number of pairs will
// give the last key a nil value.
func (b *builder) With(vs ...any) *builder {
	if len(vs) == 0 {
		return b
	}

	for i := 0; i < len(vs); i += 2 {
		k := vs[i]
		var v any

		if (i + 1) < len(vs) {
			v = vs[i+1]
		}

		b.with[k] = v
	}

	return b
}

// Debug level logging.  Whenever possible, you should add a debug category
// label to the log, as that will help your org maintain fine grained control
// of debug-level log filtering.
func (b builder) Debug(msg string) {
	b.log(lvlDebug, msg)
}

// Info is your standard info log.  You know. For information.
func (b builder) Info(msg string) {
	b.log(lvlInfo, msg)
}

// Error is an error level log.  It doesn't require an error, because there's no
// rule about needing an error to log at error level.  Or the reverse; feel free to
// add an error to your info or debug logs.  Log levels are just a fake labeling
// system, anyway.
func (b builder) Error(msg string) {
	b.log(lvlError, msg)
}

// ------------------------------------------------------------------------------------------------
// wrapper: io.writer
// ------------------------------------------------------------------------------------------------

// Writer is a wrapper that turns the logger embedded in
// the given ctx into an io.Writer.  All logs are currently
// info-level.
type Writer struct {
	Ctx context.Context
}

// Write writes to the the Writer's clogger.
func (w Writer) Write(p []byte) (int, error) {
	Ctx(w.Ctx).log(lvlInfo, string(p))
	return len(p), nil
}
