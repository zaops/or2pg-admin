package config

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// ValidationError 验证错误
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("字段 '%s': %s", e.Field, e.Message)
}

// ValidationResult 验证结果
type ValidationResult struct {
	Valid  bool
	Errors []ValidationError
}

// AddError 添加验证错误
func (vr *ValidationResult) AddError(field, message string) {
	vr.Valid = false
	vr.Errors = append(vr.Errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// Validator 配置验证器
type Validator struct{}

// NewValidator 创建新的验证器
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateConfig 验证完整配置
func (v *Validator) ValidateConfig(config *ProjectConfig) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// 验证项目信息
	v.validateProject(&config.Project, result)

	// 验证Oracle配置
	v.validateOracle(&config.Oracle, result)

	// 验证PostgreSQL配置
	v.validatePostgreSQL(&config.PostgreSQL, result)

	// 验证迁移配置
	v.validateMigration(&config.Migration, result)

	// 验证Oracle客户端配置
	v.validateOracleClient(&config.OracleClient, result)

	if result.Valid {
		logrus.Debug("配置验证通过")
	} else {
		logrus.Warnf("配置验证失败，发现 %d 个错误", len(result.Errors))
	}

	return result
}

// validateProject 验证项目信息
func (v *Validator) validateProject(project *ProjectInfo, result *ValidationResult) {
	// 验证项目名称
	if strings.TrimSpace(project.Name) == "" {
		result.AddError("project.name", "项目名称不能为空")
	} else if len(project.Name) > 100 {
		result.AddError("project.name", "项目名称长度不能超过100个字符")
	}

	// 验证版本号格式
	if project.Version != "" {
		if !v.isValidVersion(project.Version) {
			result.AddError("project.version", "版本号格式无效，应为 x.y.z 格式")
		}
	}
}

// validateOracle 验证Oracle配置
func (v *Validator) validateOracle(oracle *OracleConfig, result *ValidationResult) {
	// 验证主机地址
	if strings.TrimSpace(oracle.Host) == "" {
		result.AddError("oracle.host", "Oracle主机地址不能为空")
	} else if !v.isValidHost(oracle.Host) {
		result.AddError("oracle.host", "Oracle主机地址格式无效")
	}

	// 验证端口
	if oracle.Port <= 0 || oracle.Port > 65535 {
		result.AddError("oracle.port", "Oracle端口必须在1-65535范围内")
	}

	// 验证SID或Service Name
	if strings.TrimSpace(oracle.SID) == "" && strings.TrimSpace(oracle.Service) == "" {
		result.AddError("oracle.sid_or_service", "必须指定Oracle SID或Service Name")
	}

	// 验证用户名
	if strings.TrimSpace(oracle.Username) == "" {
		result.AddError("oracle.username", "Oracle用户名不能为空")
	}

	// 验证密码
	if strings.TrimSpace(oracle.Password) == "" {
		result.AddError("oracle.password", "Oracle密码不能为空")
	}
}

// validatePostgreSQL 验证PostgreSQL配置
func (v *Validator) validatePostgreSQL(postgres *PostgreConfig, result *ValidationResult) {
	// 验证主机地址
	if strings.TrimSpace(postgres.Host) == "" {
		result.AddError("postgresql.host", "PostgreSQL主机地址不能为空")
	} else if !v.isValidHost(postgres.Host) {
		result.AddError("postgresql.host", "PostgreSQL主机地址格式无效")
	}

	// 验证端口
	if postgres.Port <= 0 || postgres.Port > 65535 {
		result.AddError("postgresql.port", "PostgreSQL端口必须在1-65535范围内")
	}

	// 验证数据库名
	if strings.TrimSpace(postgres.Database) == "" {
		result.AddError("postgresql.database", "PostgreSQL数据库名不能为空")
	}

	// 验证用户名
	if strings.TrimSpace(postgres.Username) == "" {
		result.AddError("postgresql.username", "PostgreSQL用户名不能为空")
	}

	// 验证密码
	if strings.TrimSpace(postgres.Password) == "" {
		result.AddError("postgresql.password", "PostgreSQL密码不能为空")
	}
}

// validateMigration 验证迁移配置
func (v *Validator) validateMigration(migration *MigrationConfig, result *ValidationResult) {
	// 验证迁移类型
	if len(migration.Types) == 0 {
		result.AddError("migration.types", "至少需要指定一种迁移类型")
	} else {
		validTypes := map[string]bool{
			"TABLE": true, "VIEW": true, "SEQUENCE": true, "INDEX": true,
			"TRIGGER": true, "FUNCTION": true, "PROCEDURE": true, "PACKAGE": true,
			"TYPE": true, "GRANT": true, "TABLESPACE": true, "PARTITION": true,
			"COPY": true, "INSERT": true, "FDW": true, "QUERY": true,
		}
		for _, t := range migration.Types {
			if !validTypes[strings.ToUpper(t)] {
				result.AddError("migration.types", fmt.Sprintf("无效的迁移类型: %s", t))
			}
		}
	}

	// 验证并行作业数
	if migration.ParallelJobs <= 0 {
		result.AddError("migration.parallel_jobs", "并行作业数必须大于0")
	} else if migration.ParallelJobs > 32 {
		result.AddError("migration.parallel_jobs", "并行作业数不建议超过32")
	}

	// 验证批处理大小
	if migration.BatchSize <= 0 {
		result.AddError("migration.batch_size", "批处理大小必须大于0")
	}

	// 验证输出目录
	if strings.TrimSpace(migration.OutputDir) == "" {
		result.AddError("migration.output_dir", "输出目录不能为空")
	}

	// 验证日志级别
	validLogLevels := map[string]bool{
		"DEBUG": true, "INFO": true, "WARN": true, "ERROR": true,
	}
	if migration.LogLevel != "" && !validLogLevels[strings.ToUpper(migration.LogLevel)] {
		result.AddError("migration.log_level", "无效的日志级别，支持: DEBUG, INFO, WARN, ERROR")
	}
}

// validateOracleClient 验证Oracle客户端配置
func (v *Validator) validateOracleClient(client *OracleClientConfig, result *ValidationResult) {
	// 如果不是自动检测，验证客户端路径
	if !client.AutoDetect && strings.TrimSpace(client.Home) == "" {
		result.AddError("oracle_client.home", "未启用自动检测时，必须指定Oracle客户端路径")
	}
}

// isValidHost 验证主机地址格式
func (v *Validator) isValidHost(host string) bool {
	// 检查是否为IP地址
	if net.ParseIP(host) != nil {
		return true
	}

	// 检查是否为有效的域名
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`, host); matched {
		return true
	}

	// 检查是否为localhost
	if strings.ToLower(host) == "localhost" {
		return true
	}

	return false
}

// isValidVersion 验证版本号格式
func (v *Validator) isValidVersion(version string) bool {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return false
	}

	for _, part := range parts {
		if _, err := strconv.Atoi(part); err != nil {
			return false
		}
	}

	return true
}

// ValidateConnectionString 验证连接字符串
func (v *Validator) ValidateConnectionString(connStr string) bool {
	// 简单的连接字符串格式验证
	return strings.Contains(connStr, "host=") && strings.Contains(connStr, "port=")
}

// GetValidationSummary 获取验证结果摘要
func (v *Validator) GetValidationSummary(result *ValidationResult) string {
	if result.Valid {
		return "✅ 配置验证通过"
	}

	summary := fmt.Sprintf("❌ 配置验证失败，发现 %d 个错误:\n", len(result.Errors))
	for i, err := range result.Errors {
		summary += fmt.Sprintf("  %d. %s\n", i+1, err.Error())
	}

	return summary
}
