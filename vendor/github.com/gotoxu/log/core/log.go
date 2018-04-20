package core

// Logger 日志器接口
type Logger interface {
	Log(level Level, v ...interface{})
	Logf(level Level, format string, args ...interface{})
	Logln(level Level, v ...interface{})
	With(key string, value interface{}) Logger
	Sync() error
}
