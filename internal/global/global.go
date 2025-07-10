package global

import (
	"context"
	"github/JustGopher/Gotaxy/internal/pool"
	"go.uber.org/zap"
)

var (
	Ctx      context.Context
	Cancel   context.CancelFunc
	ConnPool *pool.Pool
	// ListenPort 服务端监听端口
	ListenPort string
	// ServerIP 服务端公网ip
	ServerIP string
	Log      *zap.SugaredLogger
)
