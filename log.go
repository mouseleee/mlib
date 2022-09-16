package mouselib

import "github.com/rs/zerolog"

type MLog struct {
	zerolog.Logger
}

var logger MLog = MLog{}
