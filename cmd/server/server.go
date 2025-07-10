package main

import (
	"context"
	"github/JustGopher/Gotaxy/internal/tunnel/serverCore/pool"

	"github/JustGopher/Gotaxy/internal/tunnel/serverCore/global"
	"github/JustGopher/Gotaxy/internal/tunnel/serverCore/shell"
)

func main() {
	global.Ctx, global.Cancel = context.WithCancel(context.Background())

	global.ConnPool = pool.NewPool()
	global.ConnPool.Set("9080", "127.0.0.1:8080")
	global.ConnPool.Set("9081", "127.0.0.1:8081")

	sh := shell.New()
	shell.RegisterCMD(sh)
	sh.Run()
}
