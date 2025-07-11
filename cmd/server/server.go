package main

import (
	"context"
	"github/JustGopher/Gotaxy/internal/global"
	"github/JustGopher/Gotaxy/internal/inits"
	"github/JustGopher/Gotaxy/internal/pool"
	"github/JustGopher/Gotaxy/internal/shell"
)

func main() {
	global.Ctx, global.Cancel = context.WithCancel(context.Background())

	global.ConnPool = pool.NewPool()
	global.ConnPool.Set("9080", "127.0.0.1:8080")
	global.ConnPool.Set("9081", "127.0.0.1:8081")

	global.DB = inits.DBInit(global.DB)
	inits.LogInit(global.Log)

	global.Config.ConfigLoad(global.DB)

	sh := shell.New()
	shell.RegisterCMD(sh)
	sh.Run()
}
