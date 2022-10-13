package mouselib_test

import (
	"testing"
	"time"

	"github.com/mouseleee/mouselib"
)

func TestDebugLoggerLevel(t *testing.T) {
	ls := []string{"debug", "info", "warn", "error", "fatal", "mewo"}

	for i, v := range ls {
		_, err := mouselib.CommandLogger(v)
		if i != 5 && err != nil {
			t.Error(err)
		}
		if i == 5 && err == nil {
			t.Fail()
		}
	}
}

func TestDebugLoggerLog(t *testing.T) {
	ls := []string{"debug", "info", "warn", "error", "fatal"}

	for i, v := range ls {
		l, err := mouselib.CommandLogger(v)
		if err != nil {
			t.Error(err)
		}
		if i == 0 {
			l.Debug().Msg("test debug")
		}
		if i == 1 {
			l.Debug().Msg("test info")
		}
		if i == 2 {
			l.Info().Msg("test warn")
		}
		if i == 3 {
			l.Warn().Msg("test error")
		}
		if i == 4 {
			l.Error().Msg("test fatal")
		}

	}
}

func TestFileLogger(t *testing.T) {
	l, err := mouselib.FileLogger("./log", 5, "debug")
	if err != nil {
		t.Error(err)
	}

	l.Debug().Str("test", "bob").Msg("test print")
	l.Info().Str("test", "bob").Msg("test print")
	l.Warn().Str("test", "bob").Msg("test print")
	time.Sleep(15 * time.Second)
	l.Error().Str("test", "bob").Msg("test print")
	l.Fatal().Str("test", "bob").Msg("test print")
}
