package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"ora2pg-admin/internal/config"
	"ora2pg-admin/internal/oracle"
	"ora2pg-admin/internal/utils"
)

var (
	checkVerbose bool
	checkConfig  string
)

// checkCmd 检查命令
var checkCmd = &cobra.Command{
	Use:   "检查",
	Short: "检查环境和连接状态",
	Long: `检查Oracle客户端环境、数据库连接状态等。

此命令提供多种检查功能，帮助您诊断和解决迁移环境中的问题：
• 环境检查：验证Oracle客户端、ora2pg工具等环境配置
• 连接测试：测试Oracle和PostgreSQL数据库连接

使用子命令指定具体的检查类型。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 如果没有提供子命令，显示帮助信息
		cmd.Help()
	},
}

// checkEnvCmd 环境检查命令
var checkEnvCmd = &cobra.Command{
	Use:   "环境",
	Short: "检查环境配置",
	Long: `检查Oracle客户端、ora2pg工具等环境配置。

此命令将检查：
• Oracle客户端安装状态和版本
• ora2pg工具可用性
• 系统环境变量配置
• 必要的依赖库

检查完成后会提供详细的环境报告和问题解决建议。`,
	Run: runCheckEnv,
}

// checkConnCmd 连接测试命令
var checkConnCmd = &cobra.Command{
	Use:   "连接",
	Short: "测试数据库连接",
	Long: `测试Oracle和PostgreSQL数据库连接。

此命令将测试：
• Oracle数据库连接和认证
• PostgreSQL数据库连接和认证
• 网络连通性和响应时间
• 数据库权限和访问性

需要先配置数据库连接信息才能进行连接测试。`,
	Run: runCheckConn,
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.AddCommand(checkEnvCmd)
	checkCmd.AddCommand(checkConnCmd)

	// 添加命令参数
	checkCmd.PersistentFlags().BoolVarP(&checkVerbose, "verbose", "v", false, "显示详细检查信息")
	checkCmd.PersistentFlags().StringVarP(&checkConfig, "config", "c", "", "指定配置文件路径")
}

// runCheckEnv 执行环境检查
func runCheckEnv(cmd *cobra.Command, args []string) {
	logger := utils.GetGlobalLogger()
	
	fmt.Println("🔍 环境检查")
	fmt.Println()

	// 1. 检查Oracle客户端
	fmt.Println("📋 Oracle客户端检查")
	fmt.Println("─────────────────────")
	
	detector := oracle.NewClientDetector()
	statusReport := detector.CheckClientStatus()
	
	fmt.Print(statusReport.GetStatusSummary())
	
	if statusReport.Status != "COMPATIBLE" {
		fmt.Println()
		if statusReport.Status == "NOT_INSTALLED" {
			guide := detector.GetInstallationGuide()
			fmt.Println("📥 安装指导:")
			fmt.Printf("下载地址: %s\n", guide.DownloadURL)
			fmt.Println("安装步骤:")
			for i, instruction := range guide.Instructions {
				fmt.Printf("  %d. %s\n", i+1, instruction)
			}
		}
	}

	// 2. 检查ora2pg工具
	fmt.Println()
	fmt.Println("📋 ora2pg工具检查")
	fmt.Println("─────────────────────")
	
	if checkOra2pgTool() {
		fmt.Println("✅ ora2pg工具: 已安装并可用")
		if checkVerbose {
			if version := getOra2pgVersion(); version != "" {
				fmt.Printf("   版本: %s\n", version)
			}
		}
	} else {
		fmt.Println("❌ ora2pg工具: 未找到")
		fmt.Println()
		fmt.Println("💡 解决建议:")
		fmt.Println("  1. 确认ora2pg已正确安装")
		fmt.Println("  2. 将ora2pg添加到PATH环境变量")
		fmt.Println("  3. 检查Perl环境是否正确配置")
	}

	// 3. 检查系统环境
	fmt.Println()
	fmt.Println("📋 系统环境检查")
	fmt.Println("─────────────────────")
	
	checkSystemEnvironment()

	// 4. 检查项目环境
	fmt.Println()
	fmt.Println("📋 项目环境检查")
	fmt.Println("─────────────────────")
	
	checkProjectEnvironment()

	// 5. 总结和建议
	fmt.Println()
	fmt.Println("📊 检查总结")
	fmt.Println("─────────────────────")
	
	provideSummaryAndSuggestions(statusReport)
	
	logger.Info("环境检查完成")
}

// runCheckConn 执行连接测试
func runCheckConn(cmd *cobra.Command, args []string) {
	logger := utils.GetGlobalLogger()
	
	fmt.Println("🔗 数据库连接测试")
	fmt.Println()

	// 1. 加载配置文件
	configPath := getConfigPath()
	if configPath == "" {
		fmt.Printf("%s\n", utils.FormatError(
			utils.ConfigErrors.FileNotFound("配置文件未找到")))
		fmt.Println()
		fmt.Println("💡 请先使用以下命令初始化项目:")
		fmt.Println("   ora2pg-admin 初始化 [项目名称]")
		fmt.Println("   ora2pg-admin 配置 数据库")
		return
	}

	manager := config.NewManager()
	if err := manager.LoadConfig(configPath); err != nil {
		fmt.Printf("%s\n", utils.FormatError(err))
		return
	}

	cfg := manager.GetConfig()

	// 2. 测试Oracle连接
	fmt.Println("📋 Oracle数据库连接测试")
	fmt.Println("─────────────────────────")
	
	tester := oracle.NewConnectionTester()
	oracleResult := tester.TestOracleConnection(&cfg.Oracle)
	
	fmt.Printf("状态: %s\n", oracleResult.Message)
	if oracleResult.Success {
		fmt.Printf("响应时间: %v\n", oracleResult.ResponseTime)
		if oracleResult.Details != "" {
			fmt.Printf("详情: %s\n", oracleResult.Details)
		}
	} else {
		if oracleResult.Error != "" {
			fmt.Printf("错误: %s\n", oracleResult.Error)
		}
		if oracleResult.Details != "" && checkVerbose {
			fmt.Printf("详细信息: %s\n", oracleResult.Details)
		}
		
		// 提供诊断信息
		fmt.Println()
		fmt.Println("🔍 连接诊断:")
		diagnostics := tester.GetConnectionDiagnostics(&cfg.Oracle)
		for _, diag := range diagnostics {
			fmt.Println(diag)
		}
	}

	// 3. 测试PostgreSQL连接
	fmt.Println()
	fmt.Println("📋 PostgreSQL数据库连接测试")
	fmt.Println("──────────────────────────")
	
	pgResult := tester.TestPostgreSQLConnection(&cfg.PostgreSQL)
	
	fmt.Printf("状态: %s\n", pgResult.Message)
	if pgResult.Success {
		fmt.Printf("响应时间: %v\n", pgResult.ResponseTime)
		if pgResult.Details != "" {
			fmt.Printf("详情: %s\n", pgResult.Details)
		}
	} else {
		if pgResult.Error != "" {
			fmt.Printf("错误: %s\n", pgResult.Error)
		}
		if pgResult.Details != "" && checkVerbose {
			fmt.Printf("详细信息: %s\n", pgResult.Details)
		}
		
		fmt.Println()
		fmt.Println("💡 解决建议:")
		fmt.Println("  1. 检查PostgreSQL服务是否运行")
		fmt.Println("  2. 验证主机名和端口是否正确")
		fmt.Println("  3. 确认用户名和密码是否正确")
		fmt.Println("  4. 检查防火墙设置")
	}

	// 4. 连接测试总结
	fmt.Println()
	fmt.Println("📊 连接测试总结")
	fmt.Println("─────────────────")
	
	if oracleResult.Success && pgResult.Success {
		fmt.Println("✅ 所有数据库连接测试通过")
		fmt.Println("🚀 您可以开始执行数据库迁移了")
		fmt.Println()
		fmt.Println("💡 下一步操作:")
		fmt.Println("   ora2pg-admin 迁移 结构    # 先迁移结构")
		fmt.Println("   ora2pg-admin 迁移 数据    # 再迁移数据")
		fmt.Println("   ora2pg-admin 迁移 全部    # 或完整迁移")
	} else {
		fmt.Println("❌ 部分连接测试失败")
		fmt.Println("🔧 请根据上述错误信息解决问题后重试")
		fmt.Println()
		fmt.Println("💡 常见解决方案:")
		fmt.Println("   1. 检查网络连接")
		fmt.Println("   2. 验证数据库服务状态")
		fmt.Println("   3. 确认连接参数配置")
		fmt.Println("   4. 检查防火墙和安全组设置")
	}
	
	logger.Info("连接测试完成")
}

// checkOra2pgTool 检查ora2pg工具
func checkOra2pgTool() bool {
	// 在PATH中查找ora2pg
	_, err := exec.LookPath("ora2pg")
	return err == nil
}

// getOra2pgVersion 获取ora2pg版本
func getOra2pgVersion() string {
	cmd := exec.Command("ora2pg", "--version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// 解析版本信息
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "ora2pg") && strings.Contains(line, "v") {
			return line
		}
	}

	return strings.TrimSpace(string(output))
}

// checkSystemEnvironment 检查系统环境
func checkSystemEnvironment() {
	// 检查ORACLE_HOME环境变量
	if oracleHome := os.Getenv("ORACLE_HOME"); oracleHome != "" {
		fmt.Printf("✅ ORACLE_HOME: %s\n", oracleHome)
	} else {
		fmt.Println("⚠️ ORACLE_HOME: 未设置")
	}

	// 检查PATH环境变量
	if path := os.Getenv("PATH"); path != "" {
		fmt.Println("✅ PATH: 已设置")
		if checkVerbose {
			fmt.Printf("   内容: %s\n", path)
		}
	} else {
		fmt.Println("❌ PATH: 未设置")
	}

	// 检查LD_LIBRARY_PATH (Linux/macOS)
	if runtime.GOOS != "windows" {
		if ldPath := os.Getenv("LD_LIBRARY_PATH"); ldPath != "" {
			fmt.Printf("✅ LD_LIBRARY_PATH: %s\n", ldPath)
		} else {
			fmt.Println("⚠️ LD_LIBRARY_PATH: 未设置")
		}
	}

	// 检查当前工作目录
	if wd, err := os.Getwd(); err == nil {
		fmt.Printf("✅ 工作目录: %s\n", wd)
	} else {
		fmt.Printf("❌ 工作目录: 获取失败 (%v)\n", err)
	}
}

// checkProjectEnvironment 检查项目环境
func checkProjectEnvironment() {
	fileUtils := utils.NewFileUtils()

	// 检查是否在项目目录中
	if fileUtils.DirExists(".ora2pg-admin") {
		fmt.Println("✅ 项目环境: 已初始化")

		// 检查配置文件
		configPath := filepath.Join(".ora2pg-admin", "config.yaml")
		if fileUtils.FileExists(configPath) {
			fmt.Println("✅ 配置文件: 存在")

			// 验证配置文件
			manager := config.NewManager()
			if err := manager.LoadConfig(configPath); err == nil {
				validator := config.NewValidator()
				result := validator.ValidateConfig(manager.GetConfig())
				if result.Valid {
					fmt.Println("✅ 配置验证: 通过")
				} else {
					fmt.Printf("⚠️ 配置验证: 发现 %d 个问题\n", len(result.Errors))
					if checkVerbose {
						for i, err := range result.Errors {
							fmt.Printf("   %d. %s\n", i+1, err.Error())
						}
					}
				}
			} else {
				fmt.Printf("❌ 配置文件: 解析失败 (%v)\n", err)
			}
		} else {
			fmt.Println("❌ 配置文件: 不存在")
		}

		// 检查输出目录
		if fileUtils.DirExists("output") {
			fmt.Println("✅ 输出目录: 存在")
		} else {
			fmt.Println("⚠️ 输出目录: 不存在")
		}

		// 检查日志目录
		if fileUtils.DirExists("logs") {
			fmt.Println("✅ 日志目录: 存在")
		} else {
			fmt.Println("⚠️ 日志目录: 不存在")
		}
	} else {
		fmt.Println("❌ 项目环境: 未初始化")
		fmt.Println()
		fmt.Println("💡 请使用以下命令初始化项目:")
		fmt.Println("   ora2pg-admin 初始化 [项目名称]")
	}
}

// provideSummaryAndSuggestions 提供总结和建议
func provideSummaryAndSuggestions(statusReport *oracle.ClientStatusReport) {
	issues := []string{}
	suggestions := []string{}

	// 检查Oracle客户端状态
	switch statusReport.Status {
	case "NOT_INSTALLED":
		issues = append(issues, "Oracle客户端未安装")
		suggestions = append(suggestions, "安装Oracle Instant Client")
	case "INCOMPATIBLE":
		issues = append(issues, "Oracle客户端版本可能不兼容")
		suggestions = append(suggestions, "升级到支持的Oracle版本")
	case "UNKNOWN_VERSION":
		issues = append(issues, "无法确定Oracle客户端版本")
		suggestions = append(suggestions, "检查Oracle客户端安装完整性")
	}

	// 检查ora2pg工具
	if !checkOra2pgTool() {
		issues = append(issues, "ora2pg工具未找到")
		suggestions = append(suggestions, "安装ora2pg工具并添加到PATH")
	}

	// 检查项目环境
	fileUtils := utils.NewFileUtils()
	if !fileUtils.DirExists(".ora2pg-admin") {
		issues = append(issues, "项目未初始化")
		suggestions = append(suggestions, "使用 'ora2pg-admin 初始化' 创建项目")
	}

	// 显示总结
	if len(issues) == 0 {
		fmt.Println("✅ 环境检查通过，所有组件正常")
		fmt.Println("🚀 您可以开始配置和执行数据库迁移")
	} else {
		fmt.Printf("⚠️ 发现 %d 个问题需要解决:\n", len(issues))
		for i, issue := range issues {
			fmt.Printf("  %d. %s\n", i+1, issue)
		}
	}

	// 显示建议
	if len(suggestions) > 0 {
		fmt.Println()
		fmt.Println("💡 解决建议:")
		for i, suggestion := range suggestions {
			fmt.Printf("  %d. %s\n", i+1, suggestion)
		}
	}

	// 显示后续步骤
	fmt.Println()
	fmt.Println("📋 推荐的操作顺序:")
	fmt.Println("  1. 解决上述环境问题")
	fmt.Println("  2. 配置数据库连接: ora2pg-admin 配置 数据库")
	fmt.Println("  3. 测试数据库连接: ora2pg-admin 检查 连接")
	fmt.Println("  4. 执行数据库迁移: ora2pg-admin 迁移 全部")
}

// getConfigPath 获取配置文件路径
func getConfigPath() string {
	// 1. 检查命令行参数指定的配置文件
	if checkConfig != "" {
		return checkConfig
	}

	// 2. 检查当前目录的项目配置
	fileUtils := utils.NewFileUtils()
	projectConfig := filepath.Join(".ora2pg-admin", "config.yaml")
	if fileUtils.FileExists(projectConfig) {
		return projectConfig
	}

	// 3. 检查可执行文件目录的配置文件
	if execPath, err := fileUtils.GetExecutablePath(); err == nil {
		execConfig := filepath.Join(execPath, ".ora2pg-admin.yaml")
		if fileUtils.FileExists(execConfig) {
			return execConfig
		}
	}

	return ""
}
