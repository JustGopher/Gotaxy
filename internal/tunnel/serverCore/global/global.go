package global

import (
	"context"
	"database/sql"
	"github/JustGopher/Gotaxy/internal/pool"
)

var (
	Ctx      context.Context
	Cancel   context.CancelFunc
	ConnPool *pool.Pool
	DB       *sql.DB
)
