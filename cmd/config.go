package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"ora2pg-admin/internal/config"
	"ora2pg-admin/internal/oracle"
	"ora2pg-admin/internal/utils"
)

var (
	configFile   string
	configBackup bool
	configForce  bool
)

// configCmd 配置命令
var configCmd = &cobra.Command{
	Use:   "配置",
	Short: "配置数据库连接和迁移选项",
	Long: `配置Oracle和PostgreSQL数据库连接，以及迁移相关选项。

此命令提供交互式配置向导，帮助您轻松配置：
• 数据库连接信息（Oracle和PostgreSQL）
• 迁移类型和选项
• 性能参数和高级设置

使用子命令指定具体的配置类型。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 如果没有提供子命令，显示帮助信息
		cmd.Help()
	},
}

// configDbCmd 数据库配置命令
var configDbCmd = &cobra.Command{
	Use:   "数据库",
	Short: "配置数据库连接",
	Long: `交互式配置Oracle和PostgreSQL数据库连接信息。

此命令将引导您配置：
• Oracle数据库连接参数（主机、端口、SID/Service、用户名、密码）
• PostgreSQL数据库连接参数（主机、端口、数据库、用户名、密码）
• 连接测试和验证

配置完成后会自动测试连接并保存配置文件。`,
	Run: runConfigDb,
}

// configOptionsCmd 迁移选项配置命令
var configOptionsCmd = &cobra.Command{
	Use:   "选项",
	Short: "配置迁移选项",
	Long: `配置数据库迁移的类型、性能参数和高级选项。

此命令将引导您配置：
• 迁移对象类型（表、视图、序列、索引等）
• 性能参数（并行作业数、批处理大小）
• 输出设置和日志级别
• 高级迁移选项

配置完成后会生成相应的ora2pg配置文件。`,
	Run: runConfigOptions,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configDbCmd)
	configCmd.AddCommand(configOptionsCmd)

	// 添加命令参数
	configCmd.PersistentFlags().StringVarP(&configFile, "file", "f", "", "指定配置文件路径")
	configCmd.PersistentFlags().BoolVar(&configBackup, "backup", true, "配置前创建备份")
	configCmd.PersistentFlags().BoolVar(&configForce, "force", false, "强制覆盖现有配置")
}

// runConfigDb 执行数据库配置
func runConfigDb(cmd *cobra.Command, args []string) {
	logger := utils.GetGlobalLogger()
	
	fmt.Println("🔧 数据库连接配置向导")
	fmt.Println()

	// 1. 加载现有配置
	manager, err := loadOrCreateConfig()
	if err != nil {
		fmt.Printf("%s\n", utils.FormatError(err))
		os.Exit(1)
	}

	cfg := manager.GetConfig()

	// 2. 配置Oracle数据库
	fmt.Println("📊 Oracle数据库配置")
	fmt.Println("─────────────────────")
	
	if err := configureOracle(&cfg.Oracle); err != nil {
		fmt.Printf("%s\n", utils.FormatError(err))
		os.Exit(1)
	}

	// 3. 配置PostgreSQL数据库
	fmt.Println()
	fmt.Println("🐘 PostgreSQL数据库配置")
	fmt.Println("──────────────────────")
	
	if err := configurePostgreSQL(&cfg.PostgreSQL); err != nil {
		fmt.Printf("%s\n", utils.FormatError(err))
		os.Exit(1)
	}

	// 4. 测试连接
	fmt.Println()
	fmt.Println("🔗 连接测试")
	fmt.Println("─────────")
	
	testConnections(cfg)

	// 5. 保存配置
	fmt.Println()
	fmt.Println("💾 保存配置")
	fmt.Println("─────────")
	
	if err := saveConfiguration(manager); err != nil {
		fmt.Printf("%s\n", utils.FormatError(err))
		os.Exit(1)
	}

	// 6. 显示配置摘要
	showConfigurationSummary(cfg)
	
	logger.Info("数据库配置完成")
}

// runConfigOptions 执行迁移选项配置
func runConfigOptions(cmd *cobra.Command, args []string) {
	logger := utils.GetGlobalLogger()
	
	fmt.Println("⚙️ 迁移选项配置向导")
	fmt.Println()

	// 1. 加载现有配置
	manager, err := loadOrCreateConfig()
	if err != nil {
		fmt.Printf("%s\n", utils.FormatError(err))
		os.Exit(1)
	}

	cfg := manager.GetConfig()

	// 2. 配置迁移类型
	fmt.Println("📋 迁移对象类型配置")
	fmt.Println("─────────────────")
	
	if err := configureMigrationTypes(&cfg.Migration); err != nil {
		fmt.Printf("%s\n", utils.FormatError(err))
		os.Exit(1)
	}

	// 3. 配置性能参数
	fmt.Println()
	fmt.Println("🚀 性能参数配置")
	fmt.Println("─────────────")
	
	if err := configurePerformanceSettings(&cfg.Migration); err != nil {
		fmt.Printf("%s\n", utils.FormatError(err))
		os.Exit(1)
	}

	// 4. 配置高级选项
	fmt.Println()
	fmt.Println("🔧 高级选项配置")
	fmt.Println("─────────────")
	
	if err := configureAdvancedOptions(&cfg.Migration); err != nil {
		fmt.Printf("%s\n", utils.FormatError(err))
		os.Exit(1)
	}

	// 5. 预览配置
	fmt.Println()
	fmt.Println("👀 配置预览")
	fmt.Println("─────────")
	
	previewMigrationConfig(&cfg.Migration)

	// 6. 确认并保存
	if confirmConfiguration() {
		if err := saveConfiguration(manager); err != nil {
			fmt.Printf("%s\n", utils.FormatError(err))
			os.Exit(1)
		}

		// 生成ora2pg配置文件
		if err := generateOra2pgConfig(cfg); err != nil {
			logger.Warnf("生成ora2pg配置文件时出现警告: %v", err)
		}

		fmt.Println()
		fmt.Println("✅ 迁移选项配置完成！")
		fmt.Println()
		fmt.Println("🚀 下一步操作:")
		fmt.Println("   ora2pg-admin 检查 连接    # 测试数据库连接")
		fmt.Println("   ora2pg-admin 迁移 全部    # 开始迁移")
	} else {
		fmt.Println("❌ 配置已取消")
	}
	
	logger.Info("迁移选项配置完成")
}

// loadOrCreateConfig 加载或创建配置
func loadOrCreateConfig() (*config.Manager, error) {
	manager := config.NewManager()

	// 确定配置文件路径
	configPath := getConfigFilePath()

	// 检查配置文件是否存在
	fileUtils := utils.NewFileUtils()
	if fileUtils.FileExists(configPath) {
		// 创建备份
		if configBackup {
			backupPath := configPath + ".backup"
			if err := fileUtils.CopyFile(configPath, backupPath); err != nil {
				return nil, utils.FileErrors.CreateFailed(backupPath, err)
			}
			fmt.Printf("📋 已创建配置备份: %s\n", backupPath)
		}

		// 加载现有配置
		if err := manager.LoadConfig(configPath); err != nil {
			return nil, utils.ConfigErrors.ParseFailed(err)
		}
		fmt.Printf("📂 已加载现有配置: %s\n", configPath)
	} else {
		// 检查是否在项目目录中
		if !fileUtils.DirExists(".ora2pg-admin") {
			return nil, utils.NewError(utils.ErrorTypeConfig, "PROJECT_NOT_INITIALIZED").
				Message("项目未初始化").
				Suggestion("请先使用 'ora2pg-admin 初始化 [项目名称]' 创建项目").
				Build()
		}

		// 创建默认配置
		manager.CreateDefaultConfig("未命名项目")
		fmt.Println("📝 已创建默认配置")
	}

	return manager, nil
}

// getConfigFilePath 获取配置文件路径
func getConfigFilePath() string {
	if configFile != "" {
		return configFile
	}

	// 默认使用项目目录的配置文件
	return filepath.Join(".ora2pg-admin", "config.yaml")
}

// configureOracle 配置Oracle数据库
func configureOracle(oracleConfig *config.OracleConfig) error {
	// 显示当前配置
	if oracleConfig.Host != "" {
		fmt.Printf("当前配置: %s:%d/%s (用户: %s)\n",
			oracleConfig.Host, oracleConfig.Port,
			getOracleIdentifier(oracleConfig), oracleConfig.Username)
		fmt.Println()
	}

	// 配置主机
	hostPrompt := promptui.Prompt{
		Label:    "Oracle主机地址",
		Default:  oracleConfig.Host,
		Validate: validateHost,
	}
	host, err := hostPrompt.Run()
	if err != nil {
		return utils.NewError(utils.ErrorTypeUser, "INPUT_CANCELLED").
			Message("用户取消了输入").Build()
	}
	oracleConfig.Host = strings.TrimSpace(host)

	// 配置端口
	portPrompt := promptui.Prompt{
		Label:    "Oracle端口",
		Default:  strconv.Itoa(oracleConfig.Port),
		Validate: validatePort,
	}
	portStr, err := portPrompt.Run()
	if err != nil {
		return utils.NewError(utils.ErrorTypeUser, "INPUT_CANCELLED").
			Message("用户取消了输入").Build()
	}
	if port, err := strconv.Atoi(strings.TrimSpace(portStr)); err == nil {
		oracleConfig.Port = port
	}

	// 选择SID或Service Name
	typePrompt := promptui.Select{
		Label: "选择Oracle连接类型",
		Items: []string{
			"SID - 系统标识符",
			"Service Name - 服务名称",
		},
	}
	typeIndex, _, err := typePrompt.Run()
	if err != nil {
		return utils.NewError(utils.ErrorTypeUser, "INPUT_CANCELLED").
			Message("用户取消了选择").Build()
	}

	if typeIndex == 0 {
		// 配置SID
		sidPrompt := promptui.Prompt{
			Label:    "Oracle SID",
			Default:  oracleConfig.SID,
			Validate: validateRequired,
		}
		sid, err := sidPrompt.Run()
		if err != nil {
			return utils.NewError(utils.ErrorTypeUser, "INPUT_CANCELLED").
				Message("用户取消了输入").Build()
		}
		oracleConfig.SID = strings.TrimSpace(sid)
		oracleConfig.Service = "" // 清空Service Name
	} else {
		// 配置Service Name
		servicePrompt := promptui.Prompt{
			Label:    "Oracle Service Name",
			Default:  oracleConfig.Service,
			Validate: validateRequired,
		}
		service, err := servicePrompt.Run()
		if err != nil {
			return utils.NewError(utils.ErrorTypeUser, "INPUT_CANCELLED").
				Message("用户取消了输入").Build()
		}
		oracleConfig.Service = strings.TrimSpace(service)
		oracleConfig.SID = "" // 清空SID
	}

	// 配置用户名
	userPrompt := promptui.Prompt{
		Label:    "Oracle用户名",
		Default:  oracleConfig.Username,
		Validate: validateRequired,
	}
	username, err := userPrompt.Run()
	if err != nil {
		return utils.NewError(utils.ErrorTypeUser, "INPUT_CANCELLED").
			Message("用户取消了输入").Build()
	}
	oracleConfig.Username = strings.TrimSpace(username)

	// 配置密码
	passwordPrompt := promptui.Prompt{
		Label: "Oracle密码",
		Mask:  '*',
		Validate: validateRequired,
	}
	password, err := passwordPrompt.Run()
	if err != nil {
		return utils.NewError(utils.ErrorTypeUser, "INPUT_CANCELLED").
			Message("用户取消了输入").Build()
	}
	oracleConfig.Password = strings.TrimSpace(password)

	// 配置Schema（可选）
	schemaPrompt := promptui.Prompt{
		Label:   "Oracle Schema（可选，直接回车跳过）",
		Default: oracleConfig.Schema,
	}
	schema, err := schemaPrompt.Run()
	if err == nil {
		oracleConfig.Schema = strings.TrimSpace(schema)
	}

	fmt.Println("✅ Oracle配置完成")
	return nil
}

// configurePostgreSQL 配置PostgreSQL数据库
func configurePostgreSQL(pgConfig *config.PostgreConfig) error {
	// 显示当前配置
	if pgConfig.Host != "" {
		fmt.Printf("当前配置: %s:%d/%s (用户: %s)\n",
			pgConfig.Host, pgConfig.Port, pgConfig.Database, pgConfig.Username)
		fmt.Println()
	}

	// 配置主机
	hostPrompt := promptui.Prompt{
		Label:    "PostgreSQL主机地址",
		Default:  pgConfig.Host,
		Validate: validateHost,
	}
	host, err := hostPrompt.Run()
	if err != nil {
		return utils.NewError(utils.ErrorTypeUser, "INPUT_CANCELLED").
			Message("用户取消了输入").Build()
	}
	pgConfig.Host = strings.TrimSpace(host)

	// 配置端口
	portPrompt := promptui.Prompt{
		Label:    "PostgreSQL端口",
		Default:  strconv.Itoa(pgConfig.Port),
		Validate: validatePort,
	}
	portStr, err := portPrompt.Run()
	if err != nil {
		return utils.NewError(utils.ErrorTypeUser, "INPUT_CANCELLED").
			Message("用户取消了输入").Build()
	}
	if port, err := strconv.Atoi(strings.TrimSpace(portStr)); err == nil {
		pgConfig.Port = port
	}

	// 配置数据库名
	dbPrompt := promptui.Prompt{
		Label:    "PostgreSQL数据库名",
		Default:  pgConfig.Database,
		Validate: validateRequired,
	}
	database, err := dbPrompt.Run()
	if err != nil {
		return utils.NewError(utils.ErrorTypeUser, "INPUT_CANCELLED").
			Message("用户取消了输入").Build()
	}
	pgConfig.Database = strings.TrimSpace(database)

	// 配置用户名
	userPrompt := promptui.Prompt{
		Label:    "PostgreSQL用户名",
		Default:  pgConfig.Username,
		Validate: validateRequired,
	}
	username, err := userPrompt.Run()
	if err != nil {
		return utils.NewError(utils.ErrorTypeUser, "INPUT_CANCELLED").
			Message("用户取消了输入").Build()
	}
	pgConfig.Username = strings.TrimSpace(username)

	// 配置密码
	passwordPrompt := promptui.Prompt{
		Label: "PostgreSQL密码",
		Mask:  '*',
		Validate: validateRequired,
	}
	password, err := passwordPrompt.Run()
	if err != nil {
		return utils.NewError(utils.ErrorTypeUser, "INPUT_CANCELLED").
			Message("用户取消了输入").Build()
	}
	pgConfig.Password = strings.TrimSpace(password)

	// 配置Schema
	schemaPrompt := promptui.Prompt{
		Label:   "PostgreSQL Schema",
		Default: pgConfig.Schema,
		Validate: validateRequired,
	}
	schema, err := schemaPrompt.Run()
	if err != nil {
		return utils.NewError(utils.ErrorTypeUser, "INPUT_CANCELLED").
			Message("用户取消了输入").Build()
	}
	pgConfig.Schema = strings.TrimSpace(schema)

	fmt.Println("✅ PostgreSQL配置完成")
	return nil
}

// 验证函数
func validateHost(input string) error {
	input = strings.TrimSpace(input)
	if input == "" {
		return fmt.Errorf("主机地址不能为空")
	}
	return nil
}

func validatePort(input string) error {
	input = strings.TrimSpace(input)
	if input == "" {
		return fmt.Errorf("端口不能为空")
	}
	port, err := strconv.Atoi(input)
	if err != nil {
		return fmt.Errorf("端口必须是数字")
	}
	if port <= 0 || port > 65535 {
		return fmt.Errorf("端口必须在1-65535范围内")
	}
	return nil
}

func validateRequired(input string) error {
	if strings.TrimSpace(input) == "" {
		return fmt.Errorf("此字段不能为空")
	}
	return nil
}

// getOracleIdentifier 获取Oracle标识符（SID或Service）
func getOracleIdentifier(oracleConfig *config.OracleConfig) string {
	if oracleConfig.Service != "" {
		return oracleConfig.Service
	}
	return oracleConfig.SID
}

// testConnections 测试数据库连接
func testConnections(cfg *config.ProjectConfig) {
	tester := oracle.NewConnectionTester()

	// 测试Oracle连接
	fmt.Print("🔍 测试Oracle连接... ")
	oracleResult := tester.TestOracleConnection(&cfg.Oracle)
	if oracleResult.Success {
		fmt.Printf("✅ 成功 (响应时间: %v)\n", oracleResult.ResponseTime)
	} else {
		fmt.Printf("❌ 失败: %s\n", oracleResult.Error)
	}

	// 测试PostgreSQL连接
	fmt.Print("🔍 测试PostgreSQL连接... ")
	pgResult := tester.TestPostgreSQLConnection(&cfg.PostgreSQL)
	if pgResult.Success {
		fmt.Printf("✅ 成功 (响应时间: %v)\n", pgResult.ResponseTime)
	} else {
		fmt.Printf("❌ 失败: %s\n", pgResult.Error)
	}

	// 显示连接测试总结
	if oracleResult.Success && pgResult.Success {
		fmt.Println("🎉 所有连接测试通过！")
	} else {
		fmt.Println("⚠️ 部分连接测试失败，请检查配置")
	}
}

// saveConfiguration 保存配置
func saveConfiguration(manager *config.Manager) error {
	configPath := getConfigFilePath()

	if err := manager.SaveConfig(configPath); err != nil {
		return utils.ConfigErrors.ParseFailed(err)
	}

	fmt.Printf("✅ 配置已保存到: %s\n", configPath)
	return nil
}

// showConfigurationSummary 显示配置摘要
func showConfigurationSummary(cfg *config.ProjectConfig) {
	fmt.Println()
	fmt.Println("📊 配置摘要")
	fmt.Println("─────────")
	fmt.Printf("Oracle:     %s:%d/%s\n", cfg.Oracle.Host, cfg.Oracle.Port, getOracleIdentifier(&cfg.Oracle))
	fmt.Printf("PostgreSQL: %s:%d/%s\n", cfg.PostgreSQL.Host, cfg.PostgreSQL.Port, cfg.PostgreSQL.Database)
	fmt.Println()
	fmt.Println("🚀 下一步操作:")
	fmt.Println("   ora2pg-admin 检查 连接    # 再次测试连接")
	fmt.Println("   ora2pg-admin 配置 选项    # 配置迁移选项")
	fmt.Println("   ora2pg-admin 迁移 全部    # 开始迁移")
}
