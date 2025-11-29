package xerr

import (
	"fmt"
	"github.com/pkg/errors"
	"runtime"
	"strconv"
)

const callerSkipOffset = 3

type withMessage struct {
	cause error
	msg   string
}

func New(err error) error {
	var path string
	f, ok := getCallerFrame(0)
	if ok {
		path = f.File + ":" + strconv.Itoa(f.Line) + ":" + err.Error()
	}
	return errors.WithMessage(err, path)
}

func WithMessage(err error, message string) error {
	if err == nil {
		return nil
	}
	var path string
	f, ok := getCallerFrame(0)
	if ok {
		path = f.File + ":" + strconv.Itoa(f.Line)
	}
	return &withMessage{
		cause: err,
		msg:   path + ":" + message,
	}
}

func WithMessagef(err error, format string, v ...any) error {
	if err == nil {
		return nil
	}
	var path string
	f, ok := getCallerFrame(0)
	if ok {
		path = f.File + ":" + strconv.Itoa(f.Line)
	}
	return &withMessage{
		cause: err,
		msg:   path + ":" + fmt.Sprintf(format, v...),
	}
}

func (w *withMessage) Error() string {
	return w.msg
}

func (w *withMessage) Cause() error {
	return w.cause
}

// 获取代码的执行行数
func getCallerFrame(skip int) (frame runtime.Frame, ok bool) {
	pc := make([]uintptr, 1)
	numFrames := runtime.Callers(skip+callerSkipOffset, pc)
	if numFrames < 1 {
		return
	}

	frame, _ = runtime.CallersFrames(pc).Next()
	return frame, frame.PC != 0
}
