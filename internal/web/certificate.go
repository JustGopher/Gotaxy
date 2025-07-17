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
)

// 证书存储目录
const certDir = "./certs"

// 标记是否需要重新生成服务端和客户端证书
var needRegenerateCerts = false

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
		global.ErrorLog.Println("生成CA证书失败: ", err)
		return
	}

	// 更新CA证书的最后修改时间
	if _, err := os.Stat(caCertPath); err == nil {
		// 如果CA证书已存在且被更新，标记需要重新生成服务端和客户端证书
		if oldCaExists {
			needRegenerateCerts = true
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if needRegenerateCerts {
		if _, err := w.Write([]byte(`{"status":"success","message":"CA证书生成成功，请重新生成服务端和客户端证书"}`)); err != nil {
			global.ErrorLog.Println("写入响应失败: ", err)
		}
	} else {
		if _, err := w.Write([]byte(`{"status":"success","message":"CA证书生成成功"}`)); err != nil {
			global.ErrorLog.Println("写入响应失败: ", err)
		}
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
	ip := global.Config.ServerIP

	// 生成服务端和客户端证书，有效期365天
	err := tlsgen.GenerateServerAndClientCerts(ip, certDir, 365, caCertPath, caKeyPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("生成证书失败: %v", err), http.StatusInternalServerError)
		global.ErrorLog.Println("生成服务端和客户端证书失败: ", err)
		return
	}

	// 重置标记，表示已经重新生成了证书
	needRegenerateCerts = false

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write([]byte(`{"status":"success","message":"服务端和客户端证书生成成功"}`)); err != nil {
		global.ErrorLog.Println("写入响应失败: ", err)
	}
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
	jsonResponse := fmt.Sprintf(`{"status":"success","data":{"caExists":%t,"certsExist":%t,"needRegenerateCerts":%t}}`,
		caExists,
		certsExist,
		needRegenerateCerts)

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write([]byte(jsonResponse)); err != nil {
		global.ErrorLog.Println("写入响应失败: ", err)
	}
}

// downloadCertsHandler 下载证书
func downloadCertsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !validateCertsForDownload(w) {
		return
	}

	// 创建临时zip文件
	zipFile, err := createTempZipFile(w)
	if err != nil {
		return
	}
	defer cleanupTempFile(zipFile)

	// 添加文件到zip
	if err := addFilesToZip(w, zipFile); err != nil {
		return
	}

	// 发送zip文件到客户端
	if err := sendZipToClient(w, zipFile); err != nil {
		return
	}
}

// validateCertsForDownload 验证下载证书前的条件
func validateCertsForDownload(w http.ResponseWriter) bool {
	// 检查所有必要的证书文件是否存在
	requiredFiles := []string{
		filepath.Join(certDir, "ca.crt"),
		filepath.Join(certDir, "client.crt"),
		filepath.Join(certDir, "client.key"),
	}

	for _, file := range requiredFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			http.Error(w, "未找到所有必要的证书文件，请确保已生成CA证书和客户端证书", http.StatusBadRequest)
			return false
		}
	}

	// 检查是否需要重新生成证书
	if needRegenerateCerts {
		http.Error(w, "CA证书已更新，请先重新生成服务端和客户端证书", http.StatusBadRequest)
		return false
	}

	return true
}

// createTempZipFile 创建临时zip文件
func createTempZipFile(w http.ResponseWriter) (*os.File, error) {
	zipFile, err := os.CreateTemp("", "certs-*.zip")
	if err != nil {
		http.Error(w, fmt.Sprintf("创建临时文件失败: %v", err), http.StatusInternalServerError)
		global.ErrorLog.Println("创建临时文件失败: ", err)
		return nil, fmt.Errorf("创建临时文件失败: %w", err)
	}
	return zipFile, nil
}

// cleanupTempFile 清理临时文件
func cleanupTempFile(zipFile *os.File) {
	if err := zipFile.Close(); err != nil {
		global.ErrorLog.Println("关闭临时文件失败: ", err)
	}
	if err := os.Remove(zipFile.Name()); err != nil {
		global.ErrorLog.Println("删除临时文件失败: ", err)
	}
}

// addFilesToZip 添加文件到zip
func addFilesToZip(w http.ResponseWriter, zipFile *os.File) error {
	// 创建zip写入器
	zipWriter := zip.NewWriter(zipFile)
	defer func() {
		if err := zipWriter.Close(); err != nil {
			global.ErrorLog.Println("关闭zip写入器失败: ", err)
		}
	}()

	// 添加文件到zip
	filesToZip := []string{"ca.crt", "client.crt", "client.key"}
	for _, filename := range filesToZip {
		filePath := filepath.Join(certDir, filename)
		fileToZip, err := os.Open(filePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("打开文件失败: %v", err), http.StatusInternalServerError)
			global.ErrorLog.Println("打开文件失败: ", err)
			return fmt.Errorf("打开文件 %s 失败: %w", filePath, err)
		}

		// 使用defer关闭文件，但需要立即执行以避免在循环结束后才关闭
		func(f *os.File) {
			defer func() {
				if err := f.Close(); err != nil {
					global.ErrorLog.Println("关闭文件失败: ", err)
				}
			}()

			// 获取文件信息
			info, err := f.Stat()
			if err != nil {
				http.Error(w, fmt.Sprintf("获取文件信息失败: %v", err), http.StatusInternalServerError)
				global.ErrorLog.Println("获取文件信息失败: ", err)
				return
			}

			// 创建zip文件头
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				http.Error(w, fmt.Sprintf("创建zip文件头失败: %v", err), http.StatusInternalServerError)
				global.ErrorLog.Println("创建zip文件头失败: ", err)
				return
			}

			// 设置压缩方法
			header.Method = zip.Deflate
			header.Name = filename

			// 创建zip文件条目
			writer, err := zipWriter.CreateHeader(header)
			if err != nil {
				http.Error(w, fmt.Sprintf("创建zip文件条目失败: %v", err), http.StatusInternalServerError)
				global.ErrorLog.Println("创建zip文件条目失败: ", err)
				return
			}

			// 复制文件内容到zip
			_, err = io.Copy(writer, f)
			if err != nil {
				http.Error(w, fmt.Sprintf("写入zip文件失败: %v", err), http.StatusInternalServerError)
				global.ErrorLog.Println("写入zip文件失败: ", err)
				return
			}
		}(fileToZip)
	}

	return nil
}

// sendZipToClient 发送zip文件到客户端
func sendZipToClient(w http.ResponseWriter, zipFile *os.File) error {
	// 关闭zip写入器
	if err := zipFile.Sync(); err != nil {
		http.Error(w, fmt.Sprintf("同步文件失败: %v", err), http.StatusInternalServerError)
		global.ErrorLog.Println("同步文件失败: ", err)
		return fmt.Errorf("同步文件失败: %w", err)
	}

	// 重置文件指针到文件开始
	if _, err := zipFile.Seek(0, 0); err != nil {
		http.Error(w, fmt.Sprintf("重置文件指针失败: %v", err), http.StatusInternalServerError)
		global.ErrorLog.Println("重置文件指针失败: ", err)
		return fmt.Errorf("重置文件指针失败: %w", err)
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=gotaxy-certs.zip")

	// 发送文件内容
	if _, err := io.Copy(w, zipFile); err != nil {
		http.Error(w, fmt.Sprintf("发送文件失败: %v", err), http.StatusInternalServerError)
		global.ErrorLog.Println("发送文件失败: ", err)
		return fmt.Errorf("发送文件失败: %w", err)
	}

	return nil
}
