package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// LogLevel 日志级别
type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
)

// LogConfig 日志配置
type LogConfig struct {
	Level      LogLevel `json:"level"`
	Format     string   `json:"format"`     // text, json
	Output     string   `json:"output"`     // stdout, stderr, file
	FilePath   string   `json:"file_path"`  // 日志文件路径
	MaxSize    int64    `json:"max_size"`   // 最大文件大小（字节）
	MaxAge     int      `json:"max_age"`    // 最大保存天数
	Compress   bool     `json:"compress"`   // 是否压缩旧日志
	TimeFormat string   `json:"time_format"` // 时间格式
}

// Logger 日志管理器
type Logger struct {
	config *LogConfig
	logger *logrus.Logger
}

// NewLogger 创建新的日志管理器
func NewLogger(config *LogConfig) *Logger {
	if config == nil {
		config = GetDefaultLogConfig()
	}

	logger := logrus.New()
	
	l := &Logger{
		config: config,
		logger: logger,
	}

	l.configure()
	return l
}

// GetDefaultLogConfig 获取默认日志配置
func GetDefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level:      LogLevelInfo,
		Format:     "text",
		Output:     "stdout",
		TimeFormat: "2006-01-02 15:04:05",
		MaxSize:    100 * 1024 * 1024, // 100MB
		MaxAge:     30,                 // 30天
		Compress:   true,
	}
}

// configure 配置日志器
func (l *Logger) configure() {
	// 设置日志级别
	l.setLevel()

	// 设置日志格式
	l.setFormatter()

	// 设置输出目标
	l.setOutput()
}

// setLevel 设置日志级别
func (l *Logger) setLevel() {
	switch l.config.Level {
	case LogLevelDebug:
		l.logger.SetLevel(logrus.DebugLevel)
	case LogLevelInfo:
		l.logger.SetLevel(logrus.InfoLevel)
	case LogLevelWarn:
		l.logger.SetLevel(logrus.WarnLevel)
	case LogLevelError:
		l.logger.SetLevel(logrus.ErrorLevel)
	default:
		l.logger.SetLevel(logrus.InfoLevel)
	}
}

// setFormatter 设置日志格式
func (l *Logger) setFormatter() {
	switch strings.ToLower(l.config.Format) {
	case "json":
		l.logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: l.config.TimeFormat,
		})
	case "text":
		fallthrough
	default:
		l.logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: l.config.TimeFormat,
			ForceColors:     l.isColorSupported(),
		})
	}
}

// setOutput 设置输出目标
func (l *Logger) setOutput() {
	switch strings.ToLower(l.config.Output) {
	case "stderr":
		l.logger.SetOutput(os.Stderr)
	case "file":
		if l.config.FilePath != "" {
			l.setFileOutput()
		} else {
			l.logger.SetOutput(os.Stdout)
		}
	case "stdout":
		fallthrough
	default:
		l.logger.SetOutput(os.Stdout)
	}
}

// setFileOutput 设置文件输出
func (l *Logger) setFileOutput() {
	// 确保日志目录存在
	logDir := filepath.Dir(l.config.FilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		l.logger.Warnf("创建日志目录失败: %v", err)
		return
	}

	// 打开日志文件
	file, err := os.OpenFile(l.config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		l.logger.Warnf("打开日志文件失败: %v", err)
		return
	}

	// 设置输出到文件
	l.logger.SetOutput(file)
}

// isColorSupported 检查是否支持颜色输出
func (l *Logger) isColorSupported() bool {
	// 在Windows命令行中通常不支持颜色，除非使用特殊终端
	// 这里简化处理，可以根据需要扩展
	return l.config.Output == "stdout" || l.config.Output == "stderr"
}

// Debug 输出调试日志
func (l *Logger) Debug(args ...interface{}) {
	l.logger.Debug(l.sanitizeMessage(fmt.Sprint(args...)))
}

// Debugf 输出格式化调试日志
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(l.sanitizeMessage(format), l.sanitizeArgs(args...)...)
}

// Info 输出信息日志
func (l *Logger) Info(args ...interface{}) {
	l.logger.Info(l.sanitizeMessage(fmt.Sprint(args...)))
}

// Infof 输出格式化信息日志
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Infof(l.sanitizeMessage(format), l.sanitizeArgs(args...)...)
}

// Warn 输出警告日志
func (l *Logger) Warn(args ...interface{}) {
	l.logger.Warn(l.sanitizeMessage(fmt.Sprint(args...)))
}

// Warnf 输出格式化警告日志
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(l.sanitizeMessage(format), l.sanitizeArgs(args...)...)
}

// Error 输出错误日志
func (l *Logger) Error(args ...interface{}) {
	l.logger.Error(l.sanitizeMessage(fmt.Sprint(args...)))
}

// Errorf 输出格式化错误日志
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(l.sanitizeMessage(format), l.sanitizeArgs(args...)...)
}

// WithField 添加字段
func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	return l.logger.WithField(key, l.sanitizeValue(value))
}

// WithFields 添加多个字段
func (l *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	sanitizedFields := make(logrus.Fields)
	for k, v := range fields {
		sanitizedFields[k] = l.sanitizeValue(v)
	}
	return l.logger.WithFields(sanitizedFields)
}

// sanitizeMessage 脱敏日志消息
func (l *Logger) sanitizeMessage(message string) string {
	// 脱敏密码相关信息
	sensitivePatterns := []string{
		"password=",
		"pwd=",
		"passwd=",
		"secret=",
		"token=",
		"key=",
	}

	result := message
	for _, pattern := range sensitivePatterns {
		if idx := strings.Index(strings.ToLower(result), pattern); idx != -1 {
			// 查找密码值的结束位置
			start := idx + len(pattern)
			end := start
			for end < len(result) && result[end] != ' ' && result[end] != ';' && result[end] != '&' && result[end] != '\n' {
				end++
			}
			// 替换为星号
			if end > start {
				result = result[:start] + strings.Repeat("*", end-start) + result[end:]
			}
		}
	}

	return result
}

// sanitizeArgs 脱敏参数
func (l *Logger) sanitizeArgs(args ...interface{}) []interface{} {
	sanitized := make([]interface{}, len(args))
	for i, arg := range args {
		sanitized[i] = l.sanitizeValue(arg)
	}
	return sanitized
}

// sanitizeValue 脱敏值
func (l *Logger) sanitizeValue(value interface{}) interface{} {
	if str, ok := value.(string); ok {
		return l.sanitizeMessage(str)
	}
	return value
}

// SetLevel 动态设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.config.Level = level
	l.setLevel()
}

// SetOutput 动态设置输出目标
func (l *Logger) SetOutput(output string, filePath ...string) {
	l.config.Output = output
	if len(filePath) > 0 {
		l.config.FilePath = filePath[0]
	}
	l.setOutput()
}

// GetLogger 获取底层logrus实例
func (l *Logger) GetLogger() *logrus.Logger {
	return l.logger
}

// Close 关闭日志器
func (l *Logger) Close() error {
	if closer, ok := l.logger.Out.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// LogRotate 日志轮转（简单实现）
func (l *Logger) LogRotate() error {
	if l.config.FilePath == "" {
		return nil
	}

	// 检查文件大小
	info, err := os.Stat(l.config.FilePath)
	if err != nil {
		return err
	}

	if info.Size() < l.config.MaxSize {
		return nil
	}

	// 生成备份文件名
	timestamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.%s", l.config.FilePath, timestamp)

	// 重命名当前日志文件
	if err := os.Rename(l.config.FilePath, backupPath); err != nil {
		return fmt.Errorf("日志轮转失败: %v", err)
	}

	// 重新设置文件输出
	l.setFileOutput()

	l.logger.Infof("日志文件已轮转: %s -> %s", l.config.FilePath, backupPath)
	return nil
}

// 全局日志实例
var globalLogger *Logger

// InitGlobalLogger 初始化全局日志器
func InitGlobalLogger(config *LogConfig) {
	globalLogger = NewLogger(config)
}

// GetGlobalLogger 获取全局日志器
func GetGlobalLogger() *Logger {
	if globalLogger == nil {
		globalLogger = NewLogger(nil)
	}
	return globalLogger
}

// 全局日志函数
func Debug(args ...interface{}) {
	GetGlobalLogger().Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	GetGlobalLogger().Debugf(format, args...)
}

func Info(args ...interface{}) {
	GetGlobalLogger().Info(args...)
}

func Infof(format string, args ...interface{}) {
	GetGlobalLogger().Infof(format, args...)
}

func Warn(args ...interface{}) {
	GetGlobalLogger().Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	GetGlobalLogger().Warnf(format, args...)
}

func Error(args ...interface{}) {
	GetGlobalLogger().Error(args...)
}

func Errorf(format string, args ...interface{}) {
	GetGlobalLogger().Errorf(format, args...)
}
