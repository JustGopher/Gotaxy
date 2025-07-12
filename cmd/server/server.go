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

	global.DB = inits.DBInit()
	inits.LogInit(global.Log)

	global.ConnPool = pool.NewPool()

	global.Config.ConfigLoad(global.DB, global.ConnPool)

	global.Log.Info("Gotaxy 启动成功")

	sh := shell.New()
	shell.RegisterCMD(sh)
	sh.Run()
}
