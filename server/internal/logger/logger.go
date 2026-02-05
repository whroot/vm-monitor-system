package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger *zap.Logger
	loggerOnce   sync.Once
	loggerConfig *Config
)

// Config 日志配置
type Config struct {
	Level      string `mapstructure:"log.level" json:"level"`
	Format      string `mapstructure:"log.format" json:"format"`
	Output      string `mapstructure:"log.output" json:"output"`
	FilePath   string `mapstructure:"log.file_path" json:"filePath"`
	MaxSize    int    `mapstructure:"log.max_size" json:"maxSize"`
	MaxBackups int    `mapstructure:"log.max_backups" json:"maxBackups"`
	MaxAge     int    `mapstructure:"log.max_age" json:"maxAge"`
	Development bool  `mapstructure:"development" json:"development"`
}

// Init 初始化日志系统
func Init(cfg *Config) error {
	loggerConfig = cfg

	var err error
	globalLogger, err = newLogger(cfg)
	if err != nil {
		return fmt.Errorf("初始化日志系统失败: %w", err)
	}

	loggerOnce.Do(func() {
		globalLogger = globalLogger.WithOptions(
			zap.AddCaller(),
			zap.AddCallerSkip(1),
		)
	})

	return nil
}

// newLogger 创建新的日志实例
func newLogger(cfg *Config) (*zap.Logger, error) {
	level, err := parseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	var zapConfig zap.Config
	if cfg.Development || cfg.Format == "development" {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zapConfig = zap.NewProductionConfig()
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	zapConfig.Level = zap.NewAtomicLevelAt(level)
	zapConfig.OutputPaths = []string{cfg.Output}
	zapConfig.ErrorOutputPaths = []string{cfg.Output}

	if cfg.Output != "stdout" && cfg.FilePath != "" {
		ensureLogDir(cfg.FilePath)
		zapConfig.OutputPaths = append(zapConfig.OutputPaths, cfg.FilePath)
		zapConfig.ErrorOutputPaths = append(zapConfig.ErrorOutputPaths, cfg.FilePath)
	}

	if cfg.MaxSize > 0 {
		zapConfig.OutputPaths = []string{"stdout"}
		if cfg.FilePath != "" {
			zapConfig.OutputPaths = append(zapConfig.OutputPaths, cfg.FilePath)
		}
	}

	return zapConfig.Build()
}

// ensureLogDir 确保日志目录存在
func ensureLogDir(filePath string) {
	dir := filepath.Dir(filePath)
	if dir == "" || dir == "." {
		return
	}
	os.MkdirAll(dir, 0755)
}

// parseLevel 解析日志级别
func parseLevel(level string) (zapcore.Level, error) {
	switch level {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn", "warning":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	case "dpanic":
		return zapcore.DPanicLevel, nil
	case "panic":
		return zapcore.PanicLevel, nil
	case "fatal":
		return zapcore.FatalLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("未知的日志级别: %s", level)
	}
}

// GetLogger 获取全局日志实例
func GetLogger() *zap.Logger {
	if globalLogger == nil {
		loggerOnce.Do(func() {
			cfg := &Config{
				Level:      "info",
				Format:     "console",
				Output:     "stdout",
				MaxSize:    100,
				MaxBackups: 10,
				MaxAge:     30,
			}
			var err error
			globalLogger, err = newLogger(cfg)
			if err != nil {
				panic(fmt.Sprintf("初始化日志系统失败: %v", err))
			}
		})
	}
	return globalLogger
}

// With 创建带有字段的日志实例
func With(fields ...zap.Field) *zap.Logger {
	return GetLogger().With(fields...)
}

// Debug 调试级别日志
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Info 信息级别日志
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Warn 警告级别日志
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error 错误级别日志
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// DPanic 致命错误日志（开发环境）
func DPanic(msg string, fields ...zap.Field) {
	GetLogger().DPanic(msg, fields...)
}

// Panic 致命错误日志
func Panic(msg string, fields ...zap.Field) {
	GetLogger().Panic(msg, fields...)
}

// Fatal 致命错误日志
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// Debugf 格式化调试日志
func Debugf(template string, args ...interface{}) {
	GetLogger().Debug(fmt.Sprintf(template, args...))
}

// Infof 格式化信息日志
func Infof(template string, args ...interface{}) {
	GetLogger().Info(fmt.Sprintf(template, args...))
}

// Warnf 格式化警告日志
func Warnf(template string, args ...interface{}) {
	GetLogger().Warn(fmt.Sprintf(template, args...))
}

// Errorf 格式化错误日志
func Errorf(template string, args ...interface{}) {
	GetLogger().Error(fmt.Sprintf(template, args...))
}

// Fatalf 格式化致命日志
func Fatalf(template string, args ...interface{}) {
	GetLogger().Fatal(fmt.Sprintf(template, args...))
}

// Sync 刷新日志缓冲区
func Sync() {
	if globalLogger != nil {
		globalLogger.Sync()
	}
}

// WithModule 创建带有模块名称的日志实例
func WithModule(module string) *zap.Logger {
	return GetLogger().With(zap.String("module", module))
}

// WithRequestID 创建带有请求ID的日志实例
func WithRequestID(requestID string) *zap.Logger {
	return GetLogger().With(zap.String("request_id", requestID))
}

// WithUserID 创建带有用户ID的日志实例
func WithUserID(userID string) *zap.Logger {
	return GetLogger().With(zap.String("user_id", userID))
}

// WithVMID 创建带有VM ID的日志实例
func WithVMID(vmID string) *zap.Logger {
	return GetLogger().With(zap.String("vm_id", vmID))
}

// LogHTTPRequest 记录HTTP请求日志
func LogHTTPRequest(method, path, clientIP, userAgent string, statusCode int, latency time.Duration) {
	GetLogger().Info("HTTP请求",
		zap.String("method", method),
		zap.String("path", path),
		zap.String("client_ip", clientIP),
		zap.String("user_agent", userAgent),
		zap.Int("status_code", statusCode),
		zap.Duration("latency", latency),
	)
}

// LogAlert 记录告警日志
func LogAlert(ruleName string, severity string, vmID string, message string) {
	GetLogger().Info("告警触发",
		zap.String("rule_name", ruleName),
		zap.String("severity", severity),
		zap.String("vm_id", vmID),
		zap.String("message", message),
	)
}

// LogNotification 记录通知日志
func LogNotification(alertID string, channel string, success bool, recipient string) {
	GetLogger().Info("通知发送",
		zap.String("alert_id", alertID),
		zap.String("channel", channel),
		zap.Bool("success", success),
		zap.String("recipient", recipient),
	)
}

// LogSync 记录同步日志
func LogSync(syncType string, added, updated, removed, failed int, duration time.Duration) {
	GetLogger().Info("VM同步完成",
		zap.String("sync_type", syncType),
		zap.Int("added", added),
		zap.Int("updated", updated),
		zap.Int("removed", removed),
		zap.Int("failed", failed),
		zap.Duration("duration", duration),
	)
}

// LogMetricCollection 记录指标采集日志
func LogMetricCollection(vmCount int, metricCount int, duration time.Duration) {
	GetLogger().Info("指标采集完成",
		zap.Int("vm_count", vmCount),
		zap.Int("metric_count", metricCount),
		zap.Duration("duration", duration),
	)
}
