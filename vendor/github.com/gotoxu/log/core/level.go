package core

// Level 表示日志的等级
type Level int8

const (
	// Debug 为调试级
	Debug Level = iota - 1

	// Info 是默认的日志等级
	Info

	// Warn 为警告级
	Warn

	// Error 为错误级
	Error

	// Panic 等级在日志记录后会触发panic
	Panic

	// Fatal 等级在日志记录后会调用os.Exit(1)
	Fatal
)
