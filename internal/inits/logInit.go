package inits

import (
	"path/filepath"

	"github/JustGopher/Gotaxy/internal/global"
	"github/JustGopher/Gotaxy/pkg/logger"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogInit 日志初始化
func LogInit(myLog *zap.SugaredLogger) {
	// 创建日志编码器（通常是 JSON 格式）
	encoder := logger.GetEncoder()
	logPath := "./logs"

	infoPath := filepath.Join(logPath, "info")
	// 获取日志输出目标（info文件）
	infoSyncer := logger.LogWriter(infoPath)
	// 将日志输出到文件
	infoCore := zapcore.NewCore(encoder, infoSyncer, zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.InfoLevel
	}))

	errorPath := filepath.Join(logPath, "error")
	// 获取日志输出目标（error文件）
	errorSyncer := logger.LogWriter(errorPath)
	errorCore := zapcore.NewCore(encoder, errorSyncer,
		zap.LevelEnablerFunc(func(level zapcore.Level) bool {
			return level >= zapcore.ErrorLevel
		}),
	)

	// 合并info文件输出和error文件输出
	core := zapcore.NewTee(infoCore, errorCore)
	log := zap.New(core, zap.AddCaller())

	global.Log = log.Sugar()
}
