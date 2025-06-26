package logger

import (
	"log/slog"
	"path/filepath"

	"github.com/mdobak/go-xerrors"
)

type stackFrame struct {
	Func   string `json:"func"`
	Source string `func:"source"`
	Line   int    `func:"line"`
}

// replaceAttr intercepts log attributes for special handling of stack trace errors
func replaceAttr(_ []string, a slog.Attr) slog.Attr {
	if a.Value.Kind() == slog.KindAny {
		if err, ok := a.Value.Any().(error); ok {
			a.Value = fmtErr(err)
		}
	}
	return a
}

// fmtErr returns a slog.Value with keys `msg` and `trace`. If the error
// does not implement interface { StackTrace() errors.StackTrace }, the `trace`
// key is omitted
func fmtErr(err error) slog.Value {
	group := []slog.Attr{
		slog.String("msg", err.Error()),
	}

	frames := marshalStack(err)
	if frames != nil {
		group = append(group, slog.Any("trace", frames))
	}

	return slog.GroupValue(group...)
}

// marshalStack extracts the error stack trace (if any)
func marshalStack(err error) []stackFrame {
	trace := xerrors.StackTrace(err)
	if len(trace) == 0 {
		return nil
	}

	frames := trace.Frames()
	result := make([]stackFrame, len(frames))

	for i, f := range frames {
		result[i] = stackFrame{
			Func: filepath.Base(f.Function),
			Source: filepath.Join(
				filepath.Base(filepath.Dir(f.File)),
				filepath.Base(f.File),
			),
			Line: f.Line,
		}
	}
	return result
}
