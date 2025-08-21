package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ora2pg-admin/internal/config"
	"ora2pg-admin/internal/oracle"
	"ora2pg-admin/internal/service"
	"ora2pg-admin/internal/utils"
)

// TestProjectInitialization 测试项目初始化流程
func TestProjectInitialization(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "ora2pg-integration-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 切换到临时目录
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// 1. 创建项目目录结构
	projectName := "集成测试项目"
	fileUtils := utils.NewFileUtils()

	projectDir := "integration_test_project"
	err = fileUtils.EnsureDir(projectDir)
	require.NoError(t, err)

	err = os.Chdir(projectDir)
	require.NoError(t, err)

	// 创建项目目录结构
	directories := []string{
		".ora2pg-admin",
		"logs",
		"output",
		"scripts",
		"backup",
		"docs",
	}

	for _, dir := range directories {
		err = fileUtils.EnsureDir(dir)
		require.NoError(t, err)
		assert.True(t, fileUtils.DirExists(dir))
	}

	// 2. 创建和验证配置文件
	manager := config.NewManager()
	manager.CreateDefaultConfig(projectName)

	configPath := filepath.Join(".ora2pg-admin", "config.yaml")
	err = manager.SaveConfig(configPath)
	require.NoError(t, err)

	// 验证配置文件存在
	assert.True(t, fileUtils.FileExists(configPath))

	// 3. 加载和验证配置
	newManager := config.NewManager()
	err = newManager.LoadConfig(configPath)
	require.NoError(t, err)

	loadedConfig := newManager.GetConfig()
	assert.Equal(t, projectName, loadedConfig.Project.Name)
	assert.Equal(t, "1.0.0", loadedConfig.Project.Version)

	// 4. 验证配置
	validator := config.NewValidator()
	result := validator.ValidateConfig(loadedConfig)
	assert.True(t, result.Valid, "配置应该是有效的")
	assert.Empty(t, result.Errors, "不应该有验证错误")

	t.Log("项目初始化集成测试通过")
}

// TestEnvironmentCheck 测试环境检查流程
func TestEnvironmentCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 1. Oracle客户端检测
	detector := oracle.NewClientDetector()
	status := detector.CheckClientStatus()

	assert.NotNil(t, status)
	assert.NotEmpty(t, status.Status)
	assert.NotEmpty(t, status.Message)

	// 状态应该是预定义的值之一
	validStatuses := []string{"COMPATIBLE", "NOT_INSTALLED", "INCOMPATIBLE", "UNKNOWN_VERSION"}
	assert.Contains(t, validStatuses, status.Status)

	// 2. 获取安装指导
	guide := detector.GetInstallationGuide()
	assert.NotNil(t, guide)
	assert.NotEmpty(t, guide.DownloadURL)
	assert.NotEmpty(t, guide.Instructions)

	// 3. 测试连接测试器
	tester := oracle.NewConnectionTester()
	assert.NotNil(t, tester)

	t.Logf("Oracle客户端状态: %s - %s", status.Status, status.Message)
	t.Log("环境检查集成测试通过")
}

// TestConfigurationWorkflow 测试配置工作流程
func TestConfigurationWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "ora2pg-config-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")

	// 1. 创建默认配置
	manager := config.NewManager()
	manager.CreateDefaultConfig("配置测试项目")

	// 2. 修改配置
	cfg := manager.GetConfig()
	cfg.Oracle.Host = "test-oracle.example.com"
	cfg.Oracle.Port = 1522
	cfg.Oracle.Username = "test_user"
	cfg.Oracle.Password = "${ORACLE_PASSWORD}"

	cfg.PostgreSQL.Host = "test-postgres.example.com"
	cfg.PostgreSQL.Port = 5433
	cfg.PostgreSQL.Username = "test_user"
	cfg.PostgreSQL.Password = "${PG_PASSWORD}"

	cfg.Migration.Types = []string{"TABLE", "VIEW", "SEQUENCE"}
	cfg.Migration.ParallelJobs = 6
	cfg.Migration.BatchSize = 2000

	// 3. 保存配置
	err = manager.SaveConfig(configPath)
	require.NoError(t, err)

	// 4. 重新加载配置
	newManager := config.NewManager()
	err = newManager.LoadConfig(configPath)
	require.NoError(t, err)

	// 5. 验证配置
	loadedConfig := newManager.GetConfig()
	assert.Equal(t, "test-oracle.example.com", loadedConfig.Oracle.Host)
	assert.Equal(t, 1522, loadedConfig.Oracle.Port)
	assert.Equal(t, "test_user", loadedConfig.Oracle.Username)
	assert.Equal(t, "${ORACLE_PASSWORD}", loadedConfig.Oracle.Password)

	assert.Equal(t, "test-postgres.example.com", loadedConfig.PostgreSQL.Host)
	assert.Equal(t, 5433, loadedConfig.PostgreSQL.Port)
	assert.Equal(t, "test_user", loadedConfig.PostgreSQL.Username)
	assert.Equal(t, "${PG_PASSWORD}", loadedConfig.PostgreSQL.Password)

	assert.Equal(t, []string{"TABLE", "VIEW", "SEQUENCE"}, loadedConfig.Migration.Types)
	assert.Equal(t, 6, loadedConfig.Migration.ParallelJobs)
	assert.Equal(t, 2000, loadedConfig.Migration.BatchSize)

	// 6. 验证配置有效性
	validator := config.NewValidator()
	result := validator.ValidateConfig(loadedConfig)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)

	t.Log("配置工作流程集成测试通过")
}

// TestServiceIntegration 测试服务集成
func TestServiceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 1. 创建ora2pg服务
	ora2pgService := service.NewOra2pgService()
	assert.NotNil(t, ora2pgService)

	// 2. 测试支持的迁移类型
	supportedTypes := ora2pgService.GetSupportedTypes()
	assert.NotEmpty(t, supportedTypes)
	assert.Contains(t, supportedTypes, service.MigrationTypeTable)
	assert.Contains(t, supportedTypes, service.MigrationTypeView)

	// 3. 测试类型验证
	for _, migrationType := range supportedTypes {
		err := ora2pgService.ValidateMigrationType(migrationType)
		assert.NoError(t, err, "类型 %s 应该是有效的", migrationType)
	}

	// 测试无效类型
	err := ora2pgService.ValidateMigrationType("INVALID_TYPE")
	assert.Error(t, err)

	// 4. 创建配置并测试迁移服务
	cfg := &config.ProjectConfig{
		Project: config.ProjectInfo{
			Name:        "集成测试项目",
			Version:     "1.0.0",
			Description: "集成测试用项目",
		},
		Oracle: config.OracleConfig{
			Host:     "localhost",
			Port:     1521,
			SID:      "ORCL",
			Username: "test_user",
			Password: "test_password",
		},
		PostgreSQL: config.PostgreConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "test_db",
			Username: "test_user",
			Password: "test_password",
			Schema:   "public",
		},
		Migration: config.MigrationConfig{
			Types:        []string{"TABLE", "VIEW"},
			ParallelJobs: 2,
			BatchSize:    1000,
			OutputDir:    "output",
			LogLevel:     "INFO",
		},
	}

	migrationService := service.NewMigrationService(cfg)
	assert.NotNil(t, migrationService)

	// 5. 测试迁移状态
	state := migrationService.GetState()
	assert.NotNil(t, state)
	assert.False(t, migrationService.IsCompleted())
	assert.False(t, migrationService.IsCancelled())
	assert.Equal(t, 0.0, migrationService.GetProgress())

	t.Log("服务集成测试通过")
}

// TestFileOperationsIntegration 测试文件操作集成
func TestFileOperationsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "ora2pg-file-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	fileUtils := utils.NewFileUtils()

	// 1. 创建复杂的目录结构
	dirs := []string{
		"project1/.ora2pg-admin",
		"project1/logs",
		"project1/output/tables",
		"project1/output/views",
		"project1/scripts/pre",
		"project1/scripts/post",
		"project1/backup",
		"project2/.ora2pg-admin",
		"project2/logs",
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(tempDir, dir)
		err = fileUtils.EnsureDir(fullPath)
		require.NoError(t, err)
		assert.True(t, fileUtils.DirExists(fullPath))
	}

	// 2. 创建和操作文件
	files := map[string]string{
		"project1/.ora2pg-admin/config.yaml": "project:\n  name: 项目1\n",
		"project1/logs/migration.log":        "迁移日志内容\n",
		"project1/scripts/pre/setup.sql":     "-- 预处理脚本\nCREATE TABLE test (id NUMBER);\n",
		"project2/.ora2pg-admin/config.yaml": "project:\n  name: 项目2\n",
	}

	for filePath, content := range files {
		fullPath := filepath.Join(tempDir, filePath)
		err = fileUtils.WriteFile(fullPath, []byte(content), 0644)
		require.NoError(t, err)
		assert.True(t, fileUtils.FileExists(fullPath))

		// 验证文件内容
		readContent, err := fileUtils.ReadFile(fullPath)
		require.NoError(t, err)
		assert.Equal(t, content, string(readContent))
	}

	// 3. 测试文件复制
	srcFile := filepath.Join(tempDir, "project1/.ora2pg-admin/config.yaml")
	dstFile := filepath.Join(tempDir, "project1/backup/config.yaml.backup")

	err = fileUtils.CopyFile(srcFile, dstFile)
	require.NoError(t, err)
	assert.True(t, fileUtils.FileExists(dstFile))

	// 验证复制的文件内容
	srcContent, err := fileUtils.ReadFile(srcFile)
	require.NoError(t, err)
	dstContent, err := fileUtils.ReadFile(dstFile)
	require.NoError(t, err)
	assert.Equal(t, srcContent, dstContent)

	// 4. 测试文件大小
	size, err := fileUtils.GetFileSize(srcFile)
	require.NoError(t, err)
	assert.Greater(t, size, int64(0))

	// 5. 测试路径操作
	fileName := fileUtils.GetFileName(srcFile)
	assert.Equal(t, "config.yaml", fileName)

	ext := fileUtils.GetFileExtension(srcFile)
	assert.Equal(t, ".yaml", ext)

	nameWithoutExt := fileUtils.GetFileNameWithoutExt(srcFile)
	assert.Equal(t, "config", nameWithoutExt)

	t.Log("文件操作集成测试通过")
}

// TestLoggerIntegration 测试日志集成
func TestLoggerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "ora2pg-log-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	logFile := filepath.Join(tempDir, "test.log")

	// 1. 创建文件日志配置
	logConfig := &utils.LogConfig{
		Level:      utils.LogLevelInfo,
		Format:     "text",
		Output:     "file",
		FilePath:   logFile,
		TimeFormat: "2006-01-02 15:04:05",
	}

	// 2. 创建日志器
	logger := utils.NewLogger(logConfig)
	assert.NotNil(t, logger)

	// 3. 写入不同级别的日志
	logger.Debug("这是调试日志（应该不显示）")
	logger.Info("这是信息日志")
	logger.Warn("这是警告日志")
	logger.Error("这是错误日志")
	logger.Infof("格式化日志: %s = %d", "测试", 42)

	// 4. 验证日志文件存在
	fileUtils := utils.NewFileUtils()
	assert.True(t, fileUtils.FileExists(logFile))

	// 5. 读取日志内容
	logContent, err := fileUtils.ReadFile(logFile)
	require.NoError(t, err)

	logStr := string(logContent)
	assert.Contains(t, logStr, "这是信息日志")
	assert.Contains(t, logStr, "这是警告日志")
	assert.Contains(t, logStr, "这是错误日志")
	assert.Contains(t, logStr, "格式化日志: 测试 = 42")
	assert.NotContains(t, logStr, "这是调试日志") // DEBUG级别不应该显示

	t.Log("日志集成测试通过")
}
