package global

import (
	"context"
	"github/JustGopher/Gotaxy/internal/tunnel/serverCore/pool"
)

var (
	Ctx      context.Context
	Cancel   context.CancelFunc
	ConnPool *pool.Pool
)
