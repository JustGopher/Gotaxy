package web

import (
	"net/http"
)

// InitRouter 初始化路由
func InitRouter() {
	// 证书生成和下载接口
	http.HandleFunc("/api/generate-ca", generateCAHandler)
	http.HandleFunc("/api/generate-certs", generateCertsHandler)
	http.HandleFunc("/api/download-certs", downloadCertsHandler)
	http.HandleFunc("/api/cert-status", certStatusHandler)
	http.HandleFunc("/api/getConf", GetConf)
	http.HandleFunc("/api/updateConf", UpdateConf)
	http.HandleFunc("/api/mappings", mappingsHandler)
	http.HandleFunc("/api/mapping/add", addMappingHandler)
	http.HandleFunc("/api/mapping/delete", delMappingHandler)
	http.HandleFunc("/api/mapping/enable", UpdateMapEna)
	http.HandleFunc("/api/start", StartService)
	http.HandleFunc("/api/stop", StopService)
	http.HandleFunc("/api/service", StatusService)
}
