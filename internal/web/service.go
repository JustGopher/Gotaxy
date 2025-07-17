package web

import (
	"context"
	"encoding/json"
	"errors"
	"github/JustGopher/Gotaxy/internal/global"
	"github/JustGopher/Gotaxy/internal/tunnel/serverCore"
	"github/JustGopher/Gotaxy/pkg/tlsgen"
	"net/http"
)

// StatusService 服务状态
func StatusService(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data": map[string]bool{
			"isRunning": global.IsRun,
		},
	})
}

// StartService 启动服务
func StartService(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if global.IsRun {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "服务已启动",
		})
		return
	}
	// 检查证书是否存在
	if !tlsgen.CheckServerCertExist("certs") {
		err := errors.New("证书缺失，请先生成证书")
		global.ErrorLog.Println("启动服务失败，err：", err)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "证书缺失，请先生成证书",
		})
		return
	}
	// 开启服务
	global.Ctx, global.Cancel = context.WithCancel(context.Background())
	go serverCore.StartServer(global.Ctx)
	global.IsRun = true
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   nil,
	})
}

// StopService 停止服务
func StopService(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !global.IsRun {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "服务已停止",
		})
		return
	}
	// 停止服务
	global.Cancel()
	global.IsRun = false
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   nil,
	})
}
