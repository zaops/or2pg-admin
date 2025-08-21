package utils

import (
	"fmt"
	"runtime"
	"strings"
)

// ErrorType 错误类型
type ErrorType string

const (
	ErrorTypeConfig     ErrorType = "CONFIG"     // 配置错误
	ErrorTypeConnection ErrorType = "CONNECTION" // 连接错误
	ErrorTypeValidation ErrorType = "VALIDATION" // 验证错误
	ErrorTypeFile       ErrorType = "FILE"       // 文件操作错误
	ErrorTypeOracle     ErrorType = "ORACLE"     // Oracle相关错误
	ErrorTypePostgres   ErrorType = "POSTGRES"   // PostgreSQL相关错误
	ErrorTypeMigration  ErrorType = "MIGRATION"  // 迁移错误
	ErrorTypeSystem     ErrorType = "SYSTEM"     // 系统错误
	ErrorTypeUser       ErrorType = "USER"       // 用户操作错误
)

// AppError 应用程序错误
type AppError struct {
	Type        ErrorType `json:"type"`
	Code        string    `json:"code"`
	Message     string    `json:"message"`
	Details     string    `json:"details,omitempty"`
	Cause       error     `json:"cause,omitempty"`
	Suggestions []string  `json:"suggestions,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	StackTrace  string    `json:"stack_trace,omitempty"`
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s:%s] %s - %s", e.Type, e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s:%s] %s", e.Type, e.Code, e.Message)
}

// Unwrap 支持errors.Unwrap
func (e *AppError) Unwrap() error {
	return e.Cause
}

// ErrorBuilder 错误构建器
type ErrorBuilder struct {
	errorType   ErrorType
	code        string
	message     string
	details     string
	cause       error
	suggestions []string
	context     map[string]interface{}
	stackTrace  bool
}

// NewError 创建新的错误构建器
func NewError(errorType ErrorType, code string) *ErrorBuilder {
	return &ErrorBuilder{
		errorType: errorType,
		code:      code,
		context:   make(map[string]interface{}),
	}
}

// Message 设置错误消息
func (eb *ErrorBuilder) Message(message string) *ErrorBuilder {
	eb.message = message
	return eb
}

// Details 设置错误详情
func (eb *ErrorBuilder) Details(details string) *ErrorBuilder {
	eb.details = details
	return eb
}

// Cause 设置错误原因
func (eb *ErrorBuilder) Cause(cause error) *ErrorBuilder {
	eb.cause = cause
	return eb
}

// Suggestion 添加解决建议
func (eb *ErrorBuilder) Suggestion(suggestion string) *ErrorBuilder {
	eb.suggestions = append(eb.suggestions, suggestion)
	return eb
}

// Suggestions 设置多个解决建议
func (eb *ErrorBuilder) Suggestions(suggestions []string) *ErrorBuilder {
	eb.suggestions = suggestions
	return eb
}

// Context 添加上下文信息
func (eb *ErrorBuilder) Context(key string, value interface{}) *ErrorBuilder {
	eb.context[key] = value
	return eb
}

// WithStackTrace 包含堆栈跟踪
func (eb *ErrorBuilder) WithStackTrace() *ErrorBuilder {
	eb.stackTrace = true
	return eb
}

// Build 构建错误
func (eb *ErrorBuilder) Build() *AppError {
	appErr := &AppError{
		Type:        eb.errorType,
		Code:        eb.code,
		Message:     eb.message,
		Details:     eb.details,
		Cause:       eb.cause,
		Suggestions: eb.suggestions,
		Context:     eb.context,
	}

	if eb.stackTrace {
		appErr.StackTrace = getStackTrace()
	}

	return appErr
}

// getStackTrace 获取堆栈跟踪
func getStackTrace() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var trace strings.Builder
	for {
		frame, more := frames.Next()
		trace.WriteString(fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function))
		if !more {
			break
		}
	}
	return trace.String()
}

// 预定义的常见错误

// ConfigErrors 配置相关错误
var ConfigErrors = struct {
	InvalidFormat    func(details string) *AppError
	MissingRequired  func(field string) *AppError
	InvalidValue     func(field, value string) *AppError
	FileNotFound     func(path string) *AppError
	ParseFailed      func(cause error) *AppError
}{
	InvalidFormat: func(details string) *AppError {
		return NewError(ErrorTypeConfig, "INVALID_FORMAT").
			Message("配置文件格式无效").
			Details(details).
			Suggestion("请检查配置文件的YAML格式是否正确").
			Suggestion("使用在线YAML验证工具检查语法").
			Build()
	},
	MissingRequired: func(field string) *AppError {
		return NewError(ErrorTypeConfig, "MISSING_REQUIRED").
			Message(fmt.Sprintf("缺少必需的配置项: %s", field)).
			Context("field", field).
			Suggestion(fmt.Sprintf("请在配置文件中添加 %s 配置项", field)).
			Build()
	},
	InvalidValue: func(field, value string) *AppError {
		return NewError(ErrorTypeConfig, "INVALID_VALUE").
			Message(fmt.Sprintf("配置项 %s 的值无效: %s", field, value)).
			Context("field", field).
			Context("value", value).
			Suggestion("请检查配置项的值是否符合要求").
			Build()
	},
	FileNotFound: func(path string) *AppError {
		return NewError(ErrorTypeConfig, "FILE_NOT_FOUND").
			Message(fmt.Sprintf("配置文件不存在: %s", path)).
			Context("path", path).
			Suggestion("请确认配置文件路径是否正确").
			Suggestion("使用 'ora2pg-admin 初始化' 命令创建新的配置文件").
			Build()
	},
	ParseFailed: func(cause error) *AppError {
		return NewError(ErrorTypeConfig, "PARSE_FAILED").
			Message("解析配置文件失败").
			Cause(cause).
			Suggestion("请检查配置文件的语法是否正确").
			Build()
	},
}

// ConnectionErrors 连接相关错误
var ConnectionErrors = struct {
	OracleClientNotFound func() *AppError
	DatabaseUnreachable  func(host string, port int) *AppError
	AuthenticationFailed func(username string) *AppError
	InvalidCredentials   func() *AppError
	TimeoutError         func() *AppError
}{
	OracleClientNotFound: func() *AppError {
		return NewError(ErrorTypeConnection, "ORACLE_CLIENT_NOT_FOUND").
			Message("未找到Oracle客户端").
			Suggestion("请安装Oracle Instant Client").
			Suggestion("设置ORACLE_HOME环境变量").
			Suggestion("将Oracle客户端路径添加到PATH环境变量").
			Build()
	},
	DatabaseUnreachable: func(host string, port int) *AppError {
		return NewError(ErrorTypeConnection, "DATABASE_UNREACHABLE").
			Message(fmt.Sprintf("无法连接到数据库 %s:%d", host, port)).
			Context("host", host).
			Context("port", port).
			Suggestion("请检查数据库服务器是否运行").
			Suggestion("验证主机名和端口是否正确").
			Suggestion("检查防火墙设置是否允许连接").
			Build()
	},
	AuthenticationFailed: func(username string) *AppError {
		return NewError(ErrorTypeConnection, "AUTHENTICATION_FAILED").
			Message(fmt.Sprintf("用户 %s 认证失败", username)).
			Context("username", username).
			Suggestion("请检查用户名和密码是否正确").
			Suggestion("确认用户账户是否被锁定").
			Build()
	},
	InvalidCredentials: func() *AppError {
		return NewError(ErrorTypeConnection, "INVALID_CREDENTIALS").
			Message("数据库凭据无效").
			Suggestion("请检查用户名和密码").
			Suggestion("确认数据库连接参数是否正确").
			Build()
	},
	TimeoutError: func() *AppError {
		return NewError(ErrorTypeConnection, "TIMEOUT").
			Message("连接超时").
			Suggestion("请检查网络连接").
			Suggestion("增加连接超时时间").
			Build()
	},
}

// FileErrors 文件操作相关错误
var FileErrors = struct {
	NotFound      func(path string) *AppError
	PermissionDenied func(path string) *AppError
	ReadFailed    func(path string, cause error) *AppError
	WriteFailed   func(path string, cause error) *AppError
	CreateFailed  func(path string, cause error) *AppError
}{
	NotFound: func(path string) *AppError {
		return NewError(ErrorTypeFile, "NOT_FOUND").
			Message(fmt.Sprintf("文件不存在: %s", path)).
			Context("path", path).
			Suggestion("请确认文件路径是否正确").
			Build()
	},
	PermissionDenied: func(path string) *AppError {
		return NewError(ErrorTypeFile, "PERMISSION_DENIED").
			Message(fmt.Sprintf("没有权限访问文件: %s", path)).
			Context("path", path).
			Suggestion("请检查文件权限设置").
			Suggestion("尝试以管理员权限运行程序").
			Build()
	},
	ReadFailed: func(path string, cause error) *AppError {
		return NewError(ErrorTypeFile, "READ_FAILED").
			Message(fmt.Sprintf("读取文件失败: %s", path)).
			Context("path", path).
			Cause(cause).
			Suggestion("请检查文件是否存在且可读").
			Build()
	},
	WriteFailed: func(path string, cause error) *AppError {
		return NewError(ErrorTypeFile, "WRITE_FAILED").
			Message(fmt.Sprintf("写入文件失败: %s", path)).
			Context("path", path).
			Cause(cause).
			Suggestion("请检查目录权限").
			Suggestion("确认磁盘空间是否充足").
			Build()
	},
	CreateFailed: func(path string, cause error) *AppError {
		return NewError(ErrorTypeFile, "CREATE_FAILED").
			Message(fmt.Sprintf("创建文件失败: %s", path)).
			Context("path", path).
			Cause(cause).
			Suggestion("请检查父目录是否存在").
			Suggestion("确认有创建文件的权限").
			Build()
	},
}

// ValidationErrors 验证相关错误
var ValidationErrors = struct {
	Required     func(field string) *AppError
	InvalidFormat func(field, format string) *AppError
	OutOfRange   func(field string, min, max interface{}) *AppError
	TooLong      func(field string, maxLength int) *AppError
	TooShort     func(field string, minLength int) *AppError
}{
	Required: func(field string) *AppError {
		return NewError(ErrorTypeValidation, "REQUIRED").
			Message(fmt.Sprintf("字段 %s 是必需的", field)).
			Context("field", field).
			Build()
	},
	InvalidFormat: func(field, format string) *AppError {
		return NewError(ErrorTypeValidation, "INVALID_FORMAT").
			Message(fmt.Sprintf("字段 %s 格式无效，期望格式: %s", field, format)).
			Context("field", field).
			Context("expected_format", format).
			Build()
	},
	OutOfRange: func(field string, min, max interface{}) *AppError {
		return NewError(ErrorTypeValidation, "OUT_OF_RANGE").
			Message(fmt.Sprintf("字段 %s 超出范围，应在 %v 到 %v 之间", field, min, max)).
			Context("field", field).
			Context("min", min).
			Context("max", max).
			Build()
	},
	TooLong: func(field string, maxLength int) *AppError {
		return NewError(ErrorTypeValidation, "TOO_LONG").
			Message(fmt.Sprintf("字段 %s 太长，最大长度为 %d", field, maxLength)).
			Context("field", field).
			Context("max_length", maxLength).
			Build()
	},
	TooShort: func(field string, minLength int) *AppError {
		return NewError(ErrorTypeValidation, "TOO_SHORT").
			Message(fmt.Sprintf("字段 %s 太短，最小长度为 %d", field, minLength)).
			Context("field", field).
			Context("min_length", minLength).
			Build()
	},
}

// FormatError 格式化错误信息用于用户显示
func FormatError(err error) string {
	if appErr, ok := err.(*AppError); ok {
		var result strings.Builder
		
		// 错误消息
		result.WriteString(fmt.Sprintf("❌ %s", appErr.Message))
		
		// 详细信息
		if appErr.Details != "" {
			result.WriteString(fmt.Sprintf("\n   详情: %s", appErr.Details))
		}
		
		// 解决建议
		if len(appErr.Suggestions) > 0 {
			result.WriteString("\n\n💡 建议:")
			for i, suggestion := range appErr.Suggestions {
				result.WriteString(fmt.Sprintf("\n   %d. %s", i+1, suggestion))
			}
		}
		
		return result.String()
	}
	
	return fmt.Sprintf("❌ %s", err.Error())
}

// IsErrorType 检查错误是否为指定类型
func IsErrorType(err error, errorType ErrorType) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == errorType
	}
	return false
}

// GetErrorCode 获取错误代码
func GetErrorCode(err error) string {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}
	return "UNKNOWN"
}
