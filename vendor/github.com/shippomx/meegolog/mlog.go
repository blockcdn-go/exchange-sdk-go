package meegolog

import (
	"os"
	"path"
	"runtime"
	"time"

	"path/filepath"

	"github.com/lestrrat/go-file-rotatelogs"
	"github.com/pkg/errors"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

var logPath string
var logFileName = "daily"

var rotationTime = time.Hour * 24
var maxAge = rotationTime * 30

func Logger() *logrus.Logger {
	logPath = "./log/"
	if !isPathExists(logPath) {
		os.Mkdir(logPath, 0777)
	}
	var logger = logrus.New()
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	file, err := os.OpenFile(logPath + "logrus.log", os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		logrus.SetOutput(file)
	} else {
		logrus.SetOutput(os.Stdout)
		logrus.Info("Failed to logger to file, using default stderr")
	}
	baseLogPath := path.Join(logPath, logFileName)
	writer, err := rotatelogs.New(
		baseLogPath+".%Y%m%d.log",
		rotatelogs.WithLinkName(baseLogPath),      // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(maxAge),             // 文件最大保存时间
		rotatelogs.WithRotationTime(rotationTime), // 日志切割时间间隔
	)
	if err != nil {
		logrus.Errorf("config local file system logger error. %+v", errors.WithStack(err))
	}
	JSONFormatter := new(logrus.JSONFormatter)
	lfHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
		logrus.PanicLevel: writer,
	}, JSONFormatter)
	logger.Formatter = customFormatter
	logger.AddHook(lfHook)
	return logger
}

func Locate(fields logrus.Fields) logrus.Fields {
	_, path, line, ok := runtime.Caller(1)
	if ok {
		_, file := filepath.Split(path)
		fields["file"] = file
		fields["line"] = line
	}
	return fields
}

func isPathExists(path string) (bool) {
    _, err := os.Stat(path)
    if err == nil {
        return true
    }
    if os.IsNotExist(err) {
        return false
    }
    return false
}
