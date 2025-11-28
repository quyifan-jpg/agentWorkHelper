/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
// Package xerr 提供了增强的错误处理功能，支持错误包装和调用栈信息追踪
package xerr

import (
	"fmt"
	"github.com/pkg/errors"
	"runtime"
	"strconv"
)

// callerSkipOffset 定义了获取调用者信息时需要跳过的栈帧数量
// 用于准确定位到实际调用错误处理函数的代码位置
const callerSkipOffset = 3

// withMessage 是一个包装错误的结构体，包含原始错误和附加消息
type withMessage struct {
	cause error  // 原始错误
	msg   string // 附加的错误消息，比如我们自己添加的记录
}

// New 创建一个包含文件路径、行号和错误信息的新错误
func New(err error) error {
	var path string
	f, ok := getCallerFrame(0)
	if ok {
		path = f.File + ":" + strconv.Itoa(f.Line) + ":" + err.Error()
	}
	return errors.WithMessage(err, path)
}

// WithMessage 创建一个包含文件路径、行号和并且包含我们自定义的附加信息的错误
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

// WithMessage 创建一个包含文件路径、行号和并且包含我们自定义的格式化附加信息的错误
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

// 返回错误里的自定义的附加信息
func (w *withMessage) Error() string {
	return w.msg
}

// 返回原始错误信息
func (w *withMessage) Cause() error {
	return w.cause
}

// getCallerFrame 获取调用者的栈帧信息，包括文件路径和行号
func getCallerFrame(skip int) (frame runtime.Frame, ok bool) {
	pc := make([]uintptr, 1)
	numFrames := runtime.Callers(skip+callerSkipOffset, pc)
	if numFrames < 1 {
		return
	}

	frame, _ = runtime.CallersFrames(pc).Next()
	return frame, frame.PC != 0
}
