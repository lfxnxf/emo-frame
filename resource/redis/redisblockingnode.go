package redis

import (
	"github.com/lfxnxf/emo-frame/logging"
	"go.uber.org/zap"

	red "github.com/go-redis/redis"
)

// ClosableNode interface represents a closable redis node.
type ClosableNode interface {
	RedisNode
	Close()
}

type (
	clientBridge struct {
		*red.Client
	}

	clusterBridge struct {
		*red.ClusterClient
	}
)

func (bridge *clientBridge) Close() {
	if err := bridge.Client.Close(); err != nil {
		logging.Errorw("Error occurred on close redis client", zap.Error(err))
	}
}

func (bridge *clusterBridge) Close() {
	if err := bridge.ClusterClient.Close(); err != nil {
		logging.Errorw("Error occurred on close redis cluster", zap.Error(err))
	}
}
