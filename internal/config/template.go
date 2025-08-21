package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/sirupsen/logrus"
)

// TemplateEngine 模板引擎
type TemplateEngine struct {
	templateDir string
}

// NewTemplateEngine 创建新的模板引擎
func NewTemplateEngine(templateDir string) *TemplateEngine {
	return &TemplateEngine{
		templateDir: templateDir,
	}
}

// GenerateOra2pgConfig 生成ora2pg配置文件
func (te *TemplateEngine) GenerateOra2pgConfig(config *ProjectConfig, outputPath string) error {
	templatePath := filepath.Join(te.templateDir, "ora2pg.conf.tmpl")
	
	// 检查模板文件是否存在
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("ora2pg配置模板文件不存在: %s", templatePath)
	}

	// 读取模板文件
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("读取模板文件失败: %v", err)
	}

	// 解析模板
	tmpl, err := template.New("ora2pg.conf").Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("解析模板失败: %v", err)
	}

	// 准备模板数据
	templateData := te.prepareOra2pgTemplateData(config)

	// 执行模板
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return fmt.Errorf("执行模板失败: %v", err)
	}

	// 确保输出目录存在
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 写入配置文件
	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("写入ora2pg配置文件失败: %v", err)
	}

	logrus.Infof("成功生成ora2pg配置文件: %s", outputPath)
	return nil
}

// GenerateProjectConfig 生成项目配置文件
func (te *TemplateEngine) GenerateProjectConfig(projectName, outputPath string) error {
	templatePath := filepath.Join(te.templateDir, "project.yaml.tmpl")
	
	// 检查模板文件是否存在
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("项目配置模板文件不存在: %s", templatePath)
	}

	// 读取模板文件
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("读取模板文件失败: %v", err)
	}

	// 解析模板
	tmpl, err := template.New("project.yaml").Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("解析模板失败: %v", err)
	}

	// 准备模板数据
	templateData := map[string]interface{}{
		"ProjectName": projectName,
		"Timestamp":   "{{.Project.Created}}",
	}

	// 执行模板
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return fmt.Errorf("执行模板失败: %v", err)
	}

	// 确保输出目录存在
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 写入配置文件
	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("写入项目配置文件失败: %v", err)
	}

	logrus.Infof("成功生成项目配置文件: %s", outputPath)
	return nil
}

// prepareOra2pgTemplateData 准备ora2pg模板数据
func (te *TemplateEngine) prepareOra2pgTemplateData(config *ProjectConfig) map[string]interface{} {
	// 构建Oracle DSN
	var oracleDSN string
	if config.Oracle.Service != "" {
		oracleDSN = fmt.Sprintf("dbi:Oracle:host=%s;service_name=%s;port=%d",
			config.Oracle.Host, config.Oracle.Service, config.Oracle.Port)
	} else {
		oracleDSN = fmt.Sprintf("dbi:Oracle:host=%s;sid=%s;port=%d",
			config.Oracle.Host, config.Oracle.SID, config.Oracle.Port)
	}

	// 构建PostgreSQL DSN
	postgreDSN := fmt.Sprintf("dbi:Pg:dbname=%s;host=%s;port=%d",
		config.PostgreSQL.Database, config.PostgreSQL.Host, config.PostgreSQL.Port)

	// 构建迁移类型字符串
	migrationTypes := ""
	for i, t := range config.Migration.Types {
		if i > 0 {
			migrationTypes += ","
		}
		migrationTypes += t
	}

	return map[string]interface{}{
		"OracleDSN":      oracleDSN,
		"OracleUser":     config.Oracle.Username,
		"OraclePassword": config.Oracle.Password,
		"OracleSchema":   config.Oracle.Schema,
		"PostgreDSN":     postgreDSN,
		"PostgreUser":    config.PostgreSQL.Username,
		"PostgrePassword": config.PostgreSQL.Password,
		"PostgreSchema":  config.PostgreSQL.Schema,
		"MigrationTypes": migrationTypes,
		"ParallelJobs":   config.Migration.ParallelJobs,
		"BatchSize":      config.Migration.BatchSize,
		"OutputDir":      config.Migration.OutputDir,
		"LogLevel":       config.Migration.LogLevel,
		"ProjectName":    config.Project.Name,
	}
}

// ValidateTemplate 验证模板文件
func (te *TemplateEngine) ValidateTemplate(templateName string) error {
	templatePath := filepath.Join(te.templateDir, templateName)
	
	// 检查模板文件是否存在
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("模板文件不存在: %s", templatePath)
	}

	// 读取模板文件
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("读取模板文件失败: %v", err)
	}

	// 尝试解析模板
	_, err = template.New(templateName).Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("模板语法错误: %v", err)
	}

	logrus.Debugf("模板文件验证通过: %s", templatePath)
	return nil
}

// GetTemplateDir 获取模板目录
func (te *TemplateEngine) GetTemplateDir() string {
	return te.templateDir
}

// SetTemplateDir 设置模板目录
func (te *TemplateEngine) SetTemplateDir(dir string) {
	te.templateDir = dir
}
