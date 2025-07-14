package web

import (
	"archive/zip"
	"fmt"
	"github/JustGopher/Gotaxy/internal/global"
	"github/JustGopher/Gotaxy/pkg/tlsgen"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// 证书存储目录
const certDir = "./certs"

// 存储CA证书的最后修改时间，用于检测CA证书是否被更新
var caLastModTime time.Time

// 标记是否需要重新生成服务端和客户端证书
var needRegenerateCerts bool = false

// InitRouter 初始化路由
func InitRouter() {
	// 证书生成和下载接口
	http.HandleFunc("/api/generate-ca", generateCAHandler)
	http.HandleFunc("/api/generate-certs", generateCertsHandler)
	http.HandleFunc("/api/download-certs", downloadCertsHandler)
	http.HandleFunc("/api/cert-status", certStatusHandler)
}

// generateCAHandler 生成根证书
func generateCAHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 检查CA证书是否已存在
	caCertPath := filepath.Join(certDir, "ca.crt")
	oldCaExists := false
	if _, err := os.Stat(caCertPath); err == nil {
		oldCaExists = true
	}

	// 生成CA证书，有效期365天，如果已存在则覆盖
	err := tlsgen.GenerateCA(certDir, 365, true)
	if err != nil {
		http.Error(w, fmt.Sprintf("生成CA证书失败: %v", err), http.StatusInternalServerError)
		global.Log.Error("生成CA证书失败: ", err)
		return
	}

	// 更新CA证书的最后修改时间
	if fileInfo, err := os.Stat(caCertPath); err == nil {
		caLastModTime = fileInfo.ModTime()
		// 如果CA证书已存在且被更新，标记需要重新生成服务端和客户端证书
		if oldCaExists {
			needRegenerateCerts = true
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if needRegenerateCerts {
		w.Write([]byte(`{"status":"success","message":"CA证书生成成功，请重新生成服务端和客户端证书"}`))
	} else {
		w.Write([]byte(`{"status":"success","message":"CA证书生成成功"}`))
	}
}

// generateCertsHandler 生成服务端和客户端证书
func generateCertsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 检查CA证书是否存在
	caCertPath := filepath.Join(certDir, "ca.crt")
	caKeyPath := filepath.Join(certDir, "ca.key")

	if _, err := os.Stat(caCertPath); os.IsNotExist(err) {
		http.Error(w, "CA证书不存在，请先生成CA证书", http.StatusBadRequest)
		return
	}

	if _, err := os.Stat(caKeyPath); os.IsNotExist(err) {
		http.Error(w, "CA密钥不存在，请先生成CA证书", http.StatusBadRequest)
		return
	}

	// 获取本机IP地址，这里简单使用127.0.0.1，实际应用中可能需要获取真实IP
	ip := "127.0.0.1"

	// 生成服务端和客户端证书，有效期365天
	err := tlsgen.GenerateServerAndClientCerts(ip, certDir, 365, caCertPath, caKeyPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("生成证书失败: %v", err), http.StatusInternalServerError)
		global.Log.Error("生成服务端和客户端证书失败: ", err)
		return
	}

	// 重置标记，表示已经重新生成了证书
	needRegenerateCerts = false

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"success","message":"服务端和客户端证书生成成功"}`))
}

// certStatusHandler 获取证书状态
func certStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 检查CA证书是否存在
	caCertPath := filepath.Join(certDir, "ca.crt")
	caExists := false
	if _, err := os.Stat(caCertPath); err == nil {
		caExists = true
	}

	// 检查服务端和客户端证书是否存在
	serverCertPath := filepath.Join(certDir, "server.crt")
	clientCertPath := filepath.Join(certDir, "client.crt")
	certsExist := false
	if _, err := os.Stat(serverCertPath); err == nil {
		if _, err := os.Stat(clientCertPath); err == nil {
			certsExist = true
		}
	}

	// 构建状态响应
	status := map[string]interface{}{
		"caExists":            caExists,
		"certsExist":          certsExist,
		"needRegenerateCerts": needRegenerateCerts,
	}

	// 转换为JSON字符串
	jsonResponse := fmt.Sprintf(`{"status":"success","data":{"caExists":%t,"certsExist":%t,"needRegenerateCerts":%t}}`,
		status["caExists"].(bool),
		status["certsExist"].(bool),
		status["needRegenerateCerts"].(bool))

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsonResponse))
}

// downloadCertsHandler 下载证书
func downloadCertsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 检查所有必要的证书文件是否存在
	requiredFiles := []string{
		filepath.Join(certDir, "ca.crt"),
		filepath.Join(certDir, "client.crt"),
		filepath.Join(certDir, "client.key"),
	}

	for _, file := range requiredFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			http.Error(w, "未找到所有必要的证书文件，请确保已生成CA证书和客户端证书", http.StatusBadRequest)
			return
		}
	}

	// 检查是否需要重新生成证书
	if needRegenerateCerts {
		http.Error(w, "CA证书已更新，请先重新生成服务端和客户端证书", http.StatusBadRequest)
		return
	}

	// 创建临时zip文件
	zipFile, err := os.CreateTemp("", "certs-*.zip")
	if err != nil {
		http.Error(w, fmt.Sprintf("创建临时文件失败: %v", err), http.StatusInternalServerError)
		global.Log.Error("创建临时文件失败: ", err)
		return
	}
	defer os.Remove(zipFile.Name())
	defer zipFile.Close()

	// 创建zip写入器
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 添加文件到zip
	filesToZip := []string{"ca.crt", "client.crt", "client.key"}
	for _, filename := range filesToZip {
		filePath := filepath.Join(certDir, filename)
		fileToZip, err := os.Open(filePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("打开文件失败: %v", err), http.StatusInternalServerError)
			global.Log.Error("打开文件失败: ", err)
			return
		}
		defer fileToZip.Close()

		// 获取文件信息
		info, err := fileToZip.Stat()
		if err != nil {
			http.Error(w, fmt.Sprintf("获取文件信息失败: %v", err), http.StatusInternalServerError)
			global.Log.Error("获取文件信息失败: ", err)
			return
		}

		// 创建zip文件头
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			http.Error(w, fmt.Sprintf("创建zip文件头失败: %v", err), http.StatusInternalServerError)
			global.Log.Error("创建zip文件头失败: ", err)
			return
		}

		// 设置压缩方法
		header.Method = zip.Deflate
		header.Name = filename

		// 创建zip文件条目
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			http.Error(w, fmt.Sprintf("创建zip文件条目失败: %v", err), http.StatusInternalServerError)
			global.Log.Error("创建zip文件条目失败: ", err)
			return
		}

		// 复制文件内容到zip
		_, err = io.Copy(writer, fileToZip)
		if err != nil {
			http.Error(w, fmt.Sprintf("写入zip文件失败: %v", err), http.StatusInternalServerError)
			global.Log.Error("写入zip文件失败: ", err)
			return
		}
	}

	// 关闭zip写入器
	zipWriter.Close()

	// 重置文件指针到文件开始
	zipFile.Seek(0, 0)

	// 设置响应头
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=gotaxy-certs.zip")

	// 发送文件内容
	_, err = io.Copy(w, zipFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("发送文件失败: %v", err), http.StatusInternalServerError)
		global.Log.Error("发送文件失败: ", err)
		return
	}
}
