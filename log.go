package mouselib

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
)

// DebugLogger 开发阶段直使用的logger
func DebugLogger() zerolog.Logger {
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

	return zerolog.New(wr).With().Timestamp().Caller().Logger().Level(zerolog.DebugLevel)
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

var levels = []string{"debug", "info", "warn", "error", "fatal"}

func NewFileLoggerWriter(filePath string) (*FileLoggerWriter, error) {
	e := make(chan error, 1)

	wr := FileLoggerWriter{}

	if _, err := os.Stat(filePath); err != nil {
		err = os.MkdirAll(filePath, os.ModeDir|0700)
		if err != nil {
			e <- nil
			return nil, err
		}
	}

	created := make([]*os.File, 0)
	defer func() {
		if <-e != nil {
			for _, v := range created {
				os.Remove(v.Name())
			}
		}
	}()

	for _, level := range levels {
		npath := filepath.Join(filePath, level+".log")
		if _, err := os.Stat(npath); err != nil {
			f, err := os.Create(npath)
			if err != nil {
				e <- err
				return nil, err
			}
			created = append(created, f)
		}

		f, err := os.OpenFile(npath, os.O_WRONLY|os.O_APPEND, 0700)
		if err != nil {
			e <- err
			return nil, err
		}
		created = append(created, f)
	}

	wr.Debug = created[0]
	wr.Info = created[1]
	wr.Warn = created[2]
	wr.Error = created[3]
	wr.Fatal = created[4]

	wr.t = time.Now()
	wr.dir = filePath

	e <- nil
	return &wr, nil
}

func (f *FileLoggerWriter) archive(ot time.Time) error {
	logger.Debug().Msg("开始归档日志...")
	suffix := fmt.Sprintf("%d%02d%02d", ot.Year(), ot.Month(), ot.Day())
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
	os.Mkdir(odir, os.ModeDir|0700)

	olds := make([]string, 0)
	for _, level := range levels {
		lpath := filepath.Join(f.dir, level+".log")
		of, _ := os.Open(lpath)

		opath := filepath.Join(odir, level+".log."+suffix)
		f, err := os.Create(opath)
		if err != nil {
			e <- err
			return err
		}

		i, err := io.Copy(f, of)
		logger.Debug().Int64("copy", i).Msg("copied")
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
	f.t = time.Now()
	logger.Debug().Msg("归档日志完成")
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
	ot := f.t
	if ct.Day()-ot.Day() >= 1 {
		err := f.archive(ot)
		if err != nil {
			logger.Err(err).Msg("归档日志发生错误")
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
	default:
		print(ori.Level)
	}
	return len(p), nil
}

func ProdLogger(filePath string, hook zerolog.Hook) (zerolog.Logger, error) {
	wr, err := NewFileLoggerWriter(filePath)
	if err != nil {
		return zerolog.Logger{}, err
	}

	multi := zerolog.MultiLevelWriter(wr, os.Stderr)

	logger := zerolog.New(multi).With().Timestamp().Logger()

	if hook != nil {
		logger = logger.Hook(hook)
	}
	return logger, nil
}
