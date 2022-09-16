package mouselib_test

import (
	"testing"
	"time"

	"github.com/mouseleee/mouselib"
)

func TestDebugLogger(t *testing.T) {
	l := mouselib.DebugLogger()

	l.Debug().Str("test", "bob").Msg("test print")
}

func TestProdLogger(t *testing.T) {
	l, err := mouselib.ProdLogger("./log", nil)
	if err != nil {
		t.Error(err)
	}

	l.Debug().Str("test", "bob").Msg("test print")
	l.Info().Str("test", "bob").Msg("test print")
	l.Warn().Str("test", "bob").Msg("test print")
	time.Sleep(4 * time.Second)
	l.Error().Str("test", "bob").Msg("test print")
	l.Fatal().Str("test", "bob").Msg("test print")
}
