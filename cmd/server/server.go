package main

import (
	"context"

	"github/JustGopher/Gotaxy/internal/tunnel/serverCore/global"
	"github/JustGopher/Gotaxy/internal/tunnel/serverCore/shell"
)

func main() {
	global.Ctx, global.Cancel = context.WithCancel(context.Background())

	// go serverCore.StartServer(global.Ctx)

	sh := shell.New()
	shell.RegisterCMD(sh)
	sh.Run()
}
