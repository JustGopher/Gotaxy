package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"time"
)

// GetEncoder 获取编码器
func GetEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}

// LogWriter 日志写入器
func LogWriter(path string) zapcore.WriteSyncer {
	// 确保目录存在
	err := os.MkdirAll(path, 0755)
	if err != nil {
		fmt.Printf("Failed to create directory %s: %v\n", path, err)
		return zapcore.AddSync(os.Stdout) // 默认输出到标准输出
	}

	// 获取当前日期
	currentDate := time.Now().Format("2006-01-02")
	fileName := fmt.Sprintf("%s.log", currentDate)
	filePath := filepath.Join(path, fileName)

	// 打开或创建文件
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open file %s: %v\n", filePath, err)
		return zapcore.AddSync(os.Stdout) // 默认输出到标准输出
	}

	return zapcore.AddSync(file)
}
