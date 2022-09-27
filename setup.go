package mouselib

import (
	"strings"

	"github.com/rs/zerolog"
)

// MouseConfig mouselib全局配置
//
// Mode 模式，debug/prod，测试和生产均使用prod
//
// LogPath 日志存储位置
type MouseConfig struct {
	Mode    string
	LogPath string
}

var logger zerolog.Logger = DebugLogger()

// SetUp 初始化库中的一些全局变量以及初始化客户端需要的参数，不显式调用则使用默认值
func SetUp(config MouseConfig) error {
	if strings.ToLower(config.Mode) == "debug" {
		logger = DebugLogger()
	}

	return nil
}
