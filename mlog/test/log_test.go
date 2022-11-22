package mlog_test

import (
	"testing"
	"time"

	"github.com/mouseleee/mlib/mlog"
	"github.com/rs/zerolog"
)

func TestInitCommandLogger(t *testing.T) {
	ls := []string{"debug", "info", "warn", "error", "fatal"}

	for _, v := range ls {
		l, err := mlog.CommandLogger(v)
		if err != nil {
			t.Error(err)
		}
		s, _ := zerolog.ParseLevel(v)
		if l.GetLevel() != s {
			t.FailNow()
		}
	}
}

func TestCommandLoggerUsage(t *testing.T) {
	ls := []string{"debug", "info", "warn", "error", "fatal"}

	for i, v := range ls {
		l, err := mlog.CommandLogger(v)
		if err != nil {
			t.Error(err)
		}
		if i == 0 {
			l.Debug().Msg("test debug")
		}
		if i == 1 {
			l.Info().Msg("test info")
		}
		if i == 2 {
			l.Warn().Msg("test warn")
		}
		if i == 3 {
			l.Error().Msg("test error")
		}
	}
}

func TestFileLogger(t *testing.T) {
	l, err := mlog.FileLogger("./log", 5, "debug")
	if err != nil {
		t.Error(err)
	}

	l.Debug().Str("test", "bob").Msg("test print")
	l.Info().Str("test", "bob").Msg("test print")
	l.Warn().Str("test", "bob").Msg("test print")
	time.Sleep(7 * time.Second)
	l.Error().Str("test", "bob").Msg("test print")
}
