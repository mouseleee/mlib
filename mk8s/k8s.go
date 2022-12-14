package mouselib

import (
	"path/filepath"

	"github.com/mouseleee/mlib/mlog"
	"github.com/rs/zerolog"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const defaultLogLevel = "debug"

var logger zerolog.Logger

func init() {
	logger, _ = mlog.CommandLogger(defaultLogLevel)
}

func initConfig(inCluster bool) (config *rest.Config, err error) {
	if inCluster {
		config, err = rest.InClusterConfig()
		if err != nil {
			logger.Err(err).Msg("使用集群内配置初始化k8s客户端失败，请检查环境")
			return nil, err
		}
	} else {
		home := homedir.HomeDir()
		k8sConfPath := filepath.Join(home, ".kube", "config")

		config, err = clientcmd.BuildConfigFromFlags("", k8sConfPath)
		if err != nil {
			p, _ := filepath.Abs(k8sConfPath)
			logger.Err(err).Str("config path", p).Msg("集群外环境使用配置文件初始化k8s客户端发生错误")
			return nil, err
		}
	}

	return config, err
}

// InitK8sClient 初始化k8s的client
//
// inCluster 是否在集群中部署
func InitK8sClient(inCluster bool) (*kubernetes.Clientset, error) {
	config, err := initConfig(inCluster)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

// InitDynamic 初始化k8s的dynamic，用户实现无结构的资源操作
//
// inCluster 是否在集群中部署
func InitDynamic(inCluster bool) (dynamic.Interface, error) {
	config, err := initConfig(inCluster)
	if err != nil {
		return nil, err
	}

	return dynamic.NewForConfig(config)
}
