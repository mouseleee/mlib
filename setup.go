package mouselib

import "github.com/rs/zerolog"

type MouseConfig struct {
	Logger zerolog.Logger
}

var logger zerolog.Logger = DebugLogger()

// SetUp 初始化库中的一些全局变量以及初始化客户端需要的参数，不显式调用则使用默认值
func SetUp(config MouseConfig) {
	logger = zerolog.Logger{}
}
