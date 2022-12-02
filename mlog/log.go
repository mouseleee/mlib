package mlog

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
)

func init() {
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05"
}

// CommandLogger 命令行logger，直接使用，默认level为DEBUG，如果level不为[debug/info/warn/error/fatal]会返回错误
func CommandLogger(level string) (zerolog.Logger, error) {
	l, err := zerolog.ParseLevel(level)
	if err != nil {
		return zerolog.Logger{}, err
	}

	wr := zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.NoColor = false
		w.TimeFormat = "2006-01-02 15:04:05"

		w.FormatMessage = func(i interface{}) string {
			return fmt.Sprintf("%s", i)
		}
		w.FormatFieldName = func(i interface{}) string {
			return fmt.Sprintf("[%s:", i)
		}
		w.FormatFieldValue = func(i interface{}) string {
			return fmt.Sprintf("%s]", i)
		}
	})

	return zerolog.New(wr).With().Timestamp().Caller().Logger().Level(l), nil
}

// FileLogger 文件logger，如果路径无效则创建logger失败，文件日志默认按天滚动
func FileLogger(filePath string, level string) (zerolog.Logger, error) {
	l, err := zerolog.ParseLevel(level)
	if err != nil {
		return zerolog.Logger{}, err
	}

	wr, err := NewFileLoggerWriter(filePath, l)
	if err != nil {
		return zerolog.Logger{}, err
	}

	return zerolog.New(wr).With().Timestamp().Caller().Logger().Level(l), nil
}

type FileLoggerWriter struct {
	Debug io.Writer
	Info  io.Writer
	Warn  io.Writer
	Error io.Writer
	Fatal io.Writer

	t   time.Time
	dir string
}

var levels = []zerolog.Level{zerolog.DebugLevel, zerolog.InfoLevel, zerolog.WarnLevel, zerolog.ErrorLevel, zerolog.FatalLevel}

func formatLevel(l zerolog.Level) string {
	switch l {
	case zerolog.DebugLevel:
		return "debug"
	case zerolog.InfoLevel:
		return "info"
	case zerolog.WarnLevel:
		return "warn"
	case zerolog.ErrorLevel:
		return "error"
	case zerolog.FatalLevel:
		return "fatal"
	}

	return ""
}

func dayZero(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}

func NewFileLoggerWriter(filePath string, level zerolog.Level) (*FileLoggerWriter, error) {
	e := make(chan error, 1)
	defer func() {
		if <-e != nil {
			err := os.RemoveAll(filePath)
			if err != nil {
				os.Exit(1)
			}
		}
	}()

	if _, err := os.Stat(filePath); err != nil {
		err = os.MkdirAll(filePath, os.ModeDir|0o700)
		if err != nil {
			e <- err
			return nil, err
		}
	}

	wr := FileLoggerWriter{}

	levls := make([]zerolog.Level, 0)
	for _, l := range levels {
		if level <= l {
			levls = append(levls, l)
		}
	}

	for _, l := range levls {
		npath := filepath.Join(filePath, formatLevel(l)+".log")
		if _, err := os.Stat(npath); err != nil {
			_, err := os.Create(npath)
			if err != nil {
				e <- err
				return nil, err
			}
		}

		f, err := os.OpenFile(npath, os.O_WRONLY|os.O_APPEND, 0700)
		if err != nil {
			e <- err
			return nil, err
		}

		if l == zerolog.DebugLevel {
			wr.Debug = f
		}
		if l == zerolog.InfoLevel {
			wr.Info = f
		}
		if l == zerolog.WarnLevel {
			wr.Warn = f
		}
		if l == zerolog.ErrorLevel {
			wr.Error = f
		}
		if l == zerolog.FatalLevel {
			wr.Fatal = f
		}
	}

	wr.t = dayZero(time.Now())
	wr.dir = filePath

	e <- nil
	return &wr, nil
}

func (f *FileLoggerWriter) archive() error {
	ot := f.t
	suffix := fmt.Sprintf("%d%02d%02d%02d%02d%02d", ot.Year(), ot.Month(), ot.Day(), ot.Hour(), ot.Minute(), ot.Second())
	created := make([]string, 0)
	e := make(chan error, 1)
	defer func() {
		if <-e != nil {
			for _, create := range created {
				os.Remove(create)
			}
		}
	}()

	odir := filepath.Join(f.dir, suffix)
	os.Mkdir(odir, os.ModeDir|0o700)

	olds := make([]string, 0)
	for _, level := range levels {
		lpath := filepath.Join(f.dir, formatLevel(level)+".log")
		of, _ := os.Open(lpath)

		opath := filepath.Join(odir, formatLevel(level)+".log."+suffix)
		f, err := os.Create(opath)
		if err != nil {
			e <- err
			return err
		}

		_, err = io.Copy(f, of)
		if err != nil {
			e <- err
			return err
		}

		olds = append(olds, lpath)
		created = append(created, f.Name())
	}

	for _, old := range olds {
		os.Create(old)
	}

	e <- nil
	f.t = dayZero(time.Now())
	return nil
}

func (f *FileLoggerWriter) Write(p []byte) (n int, err error) {
	type t struct {
		Level     string `json:"level"`
		TimeStamp string `json:"time"`
	}
	var ori t
	json.Unmarshal(p, &ori)

	ct, _ := time.Parse(time.RFC3339, ori.TimeStamp)

	if ct.Unix()-f.t.Unix() >= 86400 {
		err := f.archive()
		if err != nil {
			return 0, errors.New("归档日志错误")
		}
	}

	switch ori.Level {
	case "debug":
		f.Debug.Write(p)
	case "info":
		f.Info.Write(p)
	case "warn":
		f.Warn.Write(p)
	case "error":
		f.Error.Write(p)
	case "fatal":
		f.Fatal.Write(p)
	}

	return len(p), nil
}
