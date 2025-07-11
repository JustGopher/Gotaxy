package main

import (
	"context"
	"github/JustGopher/Gotaxy/internal/global"
	"github/JustGopher/Gotaxy/internal/pool"
	shell2 "github/JustGopher/Gotaxy/internal/shell"
)

func main() {
	global.Ctx, global.Cancel = context.WithCancel(context.Background())

	global.ConnPool = pool.NewPool()
	global.ConnPool.Set("9080", "127.0.0.1:8080")
	global.ConnPool.Set("9081", "127.0.0.1:8081")

	global.ListenPort = "9000"
	global.ServerIP = "127.0.0.1"

	sh := shell2.New()
	shell2.RegisterCMD(sh)
	sh.Run()
}
