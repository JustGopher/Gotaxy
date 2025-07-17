package web

import (
	"context"
	"errors"
	"github/JustGopher/Gotaxy/internal/global"
	"github/JustGopher/Gotaxy/internal/tunnel/serverCore"
	"github/JustGopher/Gotaxy/pkg/tlsgen"
	"net/http"
)

// StartService 启动服务
func StartService(w http.ResponseWriter, r *http.Request) {
	// 检查证书是否存在
	if !tlsgen.CheckServerCertExist("certs") {
		err := errors.New("证书缺失，请先生成证书")
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	global.Ctx, global.Cancel = context.WithCancel(context.Background())
	go serverCore.StartServer(global.Ctx)
	_, _ = w.Write([]byte("服务已启动"))
}

// StopService 停止服务
func StopService(w http.ResponseWriter, r *http.Request) {
	global.Cancel()
	_, _ = w.Write([]byte("服务已停止"))
}
