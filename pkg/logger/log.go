package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// RollingFileWriter 滚动日志写入器
type RollingFileWriter struct {
	dirPath     string
	currentDate string
	currentFile *os.File
}

// NewRollingFileWriter 创建一个新的滚动日志写入器
func NewRollingFileWriter(path string) *RollingFileWriter {
	writer := &RollingFileWriter{
		dirPath:     path,
		currentDate: time.Now().Format("2006-01-02"),
	}
	writer.ensureDir()
	return writer
}

// Write 写入日志，实现io的Writer接口
func (w *RollingFileWriter) Write(p []byte) (n int, err error) {
	// 当日期发生变化时滚动日志
	if w.shouldRotate() || w.currentFile == nil {
		w.rotate()
	}
	return w.currentFile.Write(p)
}

// ensureDir 确保目录存在
func (w *RollingFileWriter) ensureDir() {
	err := os.MkdirAll(w.dirPath, 0755)
	if err != nil {
		fmt.Printf("Failed to create directory %s: %v\n", w.dirPath, err)
	}
}

// openNewFile 打开新的日志文件
func (w *RollingFileWriter) openNewFile() {
	today := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("%s.log", today)
	filePath := filepath.Join(w.dirPath, filename)

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open file %s: %v\n", filePath, err)
		w.currentFile = os.Stdout
		return
	}

	w.currentFile = file
	w.currentDate = today
}

// shouldRotate 判断是否需要滚动
func (w *RollingFileWriter) shouldRotate() bool {
	return time.Now().Format("2006-01-02") != w.currentDate
}

// rotate 滚动日志
func (w *RollingFileWriter) rotate() {
	if w.currentFile != nil {
		_ = w.currentFile.Close()
	}
	w.openNewFile()
}
