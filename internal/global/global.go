package global

import (
	"context"
	"database/sql"
	"github/JustGopher/Gotaxy/internal/config"
	"github/JustGopher/Gotaxy/internal/pool"
	"log"
)

var (
	Ctx      context.Context
	Cancel   context.CancelFunc
	ConnPool *pool.Pool
	InfoLog  *log.Logger
	ErrorLog *log.Logger
	DB       *sql.DB
	Config   config.Config
)
