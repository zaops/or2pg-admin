package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// ProjectConfig 项目配置结构
type ProjectConfig struct {
	Project    ProjectInfo    `yaml:"project" json:"project"`
	Oracle     OracleConfig   `yaml:"oracle" json:"oracle"`
	PostgreSQL PostgreConfig  `yaml:"postgresql" json:"postgresql"`
	Migration  MigrationConfig `yaml:"migration" json:"migration"`
	OracleClient OracleClientConfig `yaml:"oracle_client" json:"oracle_client"`
}

// ProjectInfo 项目基本信息
type ProjectInfo struct {
	Name        string    `yaml:"name" json:"name"`
	Version     string    `yaml:"version" json:"version"`
	Description string    `yaml:"description" json:"description"`
	Created     time.Time `yaml:"created" json:"created"`
	Updated     time.Time `yaml:"updated" json:"updated"`
}

// OracleConfig Oracle数据库配置
type OracleConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	SID      string `yaml:"sid" json:"sid"`
	Service  string `yaml:"service" json:"service"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	Schema   string `yaml:"schema" json:"schema"`
}

// PostgreConfig PostgreSQL数据库配置
type PostgreConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Database string `yaml:"database" json:"database"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	Schema   string `yaml:"schema" json:"schema"`
}

// MigrationConfig 迁移配置
type MigrationConfig struct {
	Types        []string `yaml:"types" json:"types"`
	ParallelJobs int      `yaml:"parallel_jobs" json:"parallel_jobs"`
	BatchSize    int      `yaml:"batch_size" json:"batch_size"`
	OutputDir    string   `yaml:"output_dir" json:"output_dir"`
	LogLevel     string   `yaml:"log_level" json:"log_level"`
}

// OracleClientConfig Oracle客户端配置
type OracleClientConfig struct {
	Home       string `yaml:"home" json:"home"`
	AutoDetect bool   `yaml:"auto_detect" json:"auto_detect"`
}

// Manager 配置管理器
type Manager struct {
	config     *ProjectConfig
	configPath string
}

// NewManager 创建新的配置管理器
func NewManager() *Manager {
	return &Manager{
		config: &ProjectConfig{},
	}
}

// LoadConfig 加载配置文件
func (m *Manager) LoadConfig(configPath string) error {
	m.configPath = configPath
	
	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logrus.Debugf("配置文件不存在: %s", configPath)
		return fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析YAML配置
	if err := yaml.Unmarshal(data, m.config); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 处理环境变量替换
	m.processEnvVars()

	logrus.Infof("成功加载配置文件: %s", configPath)
	return nil
}

// SaveConfig 保存配置文件
func (m *Manager) SaveConfig(configPath string) error {
	if configPath != "" {
		m.configPath = configPath
	}

	// 确保目录存在
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 更新时间戳
	m.config.Project.Updated = time.Now()

	// 序列化为YAML
	data, err := yaml.Marshal(m.config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	logrus.Infof("成功保存配置文件: %s", m.configPath)
	return nil
}

// GetConfig 获取配置
func (m *Manager) GetConfig() *ProjectConfig {
	return m.config
}

// SetConfig 设置配置
func (m *Manager) SetConfig(config *ProjectConfig) {
	m.config = config
}

// processEnvVars 处理环境变量替换
func (m *Manager) processEnvVars() {
	// Oracle密码
	if strings.HasPrefix(m.config.Oracle.Password, "${") && strings.HasSuffix(m.config.Oracle.Password, "}") {
		envVar := strings.TrimSuffix(strings.TrimPrefix(m.config.Oracle.Password, "${"), "}")
		if value := os.Getenv(envVar); value != "" {
			m.config.Oracle.Password = value
		}
	}

	// PostgreSQL密码
	if strings.HasPrefix(m.config.PostgreSQL.Password, "${") && strings.HasSuffix(m.config.PostgreSQL.Password, "}") {
		envVar := strings.TrimSuffix(strings.TrimPrefix(m.config.PostgreSQL.Password, "${"), "}")
		if value := os.Getenv(envVar); value != "" {
			m.config.PostgreSQL.Password = value
		}
	}
}

// CreateDefaultConfig 创建默认配置
func (m *Manager) CreateDefaultConfig(projectName string) {
	m.config = &ProjectConfig{
		Project: ProjectInfo{
			Name:        projectName,
			Version:     "1.0.0",
			Description: "Oracle到PostgreSQL数据库迁移项目",
			Created:     time.Now(),
			Updated:     time.Now(),
		},
		Oracle: OracleConfig{
			Host:     "localhost",
			Port:     1521,
			SID:      "ORCL",
			Username: "system",
			Password: "${ORACLE_PASSWORD}",
			Schema:   "",
		},
		PostgreSQL: PostgreConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "postgres",
			Username: "postgres",
			Password: "${PG_PASSWORD}",
			Schema:   "public",
		},
		Migration: MigrationConfig{
			Types:        []string{"TABLE", "VIEW", "SEQUENCE", "INDEX"},
			ParallelJobs: 4,
			BatchSize:    1000,
			OutputDir:    "output",
			LogLevel:     "INFO",
		},
		OracleClient: OracleClientConfig{
			AutoDetect: true,
		},
	}
}

// GetConfigPath 获取配置文件路径
func (m *Manager) GetConfigPath() string {
	return m.configPath
}

// LoadFromViper 从Viper加载配置
func (m *Manager) LoadFromViper() error {
	if err := viper.Unmarshal(m.config); err != nil {
		return fmt.Errorf("从Viper加载配置失败: %v", err)
	}
	return nil
}
