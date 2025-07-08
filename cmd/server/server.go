package main

import (
	shellCore "github/JustGopher/Gotaxy/internal/tunnel/serverCore/shell"
	"github/JustGopher/Gotaxy/pkg/shellcli"
)

func main() {
	shell := shellcli.New()
	shellCore.Register(shell)
	shell.Run()
}
