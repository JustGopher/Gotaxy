package main

import (
	"context"
	"github/JustGopher/Gotaxy/internal/web"

	"github/JustGopher/Gotaxy/internal/global"
	"github/JustGopher/Gotaxy/internal/inits"
	"github/JustGopher/Gotaxy/internal/pool"
	"github/JustGopher/Gotaxy/internal/shell"
)

func main() {
	global.Ctx, global.Cancel = context.WithCancel(context.Background())

	global.DB = inits.DBInit()
	inits.LogInit()

	global.ConnPool = pool.NewPool()

	global.Config.ConfigLoad(global.DB, global.ConnPool)

	global.InfoLog.Println("Gotaxy 启动成功")

	go web.Start()

	sh := shell.New()
	shell.RegisterCMD(sh)
	sh.Run()
}
