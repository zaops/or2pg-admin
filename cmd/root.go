package cmd

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"ora2pg-admin/internal/utils"
)

var (
	cfgFile string
	verbose bool
	quiet   bool
	dryRun  bool
	logFile string
)

// 版本信息
var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
)

// rootCmd 代表没有调用子命令时的基础命令
var rootCmd = &cobra.Command{
	Use:   "ora2pg-admin",
	Short: "Ora2Pg 中文CLI管理器",
	Long: `Ora2Pg 中文CLI管理器是一个友好的命令行工具，用于简化Oracle到PostgreSQL数据库迁移操作。

本工具为ora2pg提供了直观的中文命令界面，让运维人员能够轻松完成数据库迁移任务，
无需学习复杂的ora2pg命令行参数。

主要功能：
• 中文命令界面，降低学习成本
• 自动生成ora2pg配置文件
• Oracle客户端环境检测
• 交互式配置向导
• 实时迁移进度跟踪`,
	Run: func(cmd *cobra.Command, args []string) {
		// 如果没有提供子命令，显示帮助信息
		cmd.Help()
	},
}

// Execute 添加所有子命令到根命令并设置适当的标志
// 这由main.main()调用。它只需要对rootCmd调用一次。
func Execute() error {
	return rootCmd.Execute()
}

// SetVersionInfo 设置版本信息
func SetVersionInfo(v, bt, gc string) {
	version = v
	buildTime = bt
	gitCommit = gc
}

func init() {
	cobra.OnInitialize(initConfig)

	// 全局标志
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径 (默认为二进制文件同目录下的 .ora2pg-admin.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "显示详细输出")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "静默模式")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "预览模式，不执行实际操作")
	rootCmd.PersistentFlags().StringVar(&logFile, "log-file", "", "指定日志文件路径")

	// 将标志绑定到viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	viper.BindPFlag("dry-run", rootCmd.PersistentFlags().Lookup("dry-run"))
	viper.BindPFlag("log-file", rootCmd.PersistentFlags().Lookup("log-file"))
}

// initConfig 读取配置文件和环境变量
func initConfig() {
	// 初始化日志系统
	initLogger()

	if cfgFile != "" {
		// 使用命令行指定的配置文件
		viper.SetConfigFile(cfgFile)
	} else {
		// 获取可执行文件的目录
		execPath, err := os.Executable()
		if err != nil {
			logrus.Warnf("无法获取可执行文件路径: %v", err)
			// 回退到当前目录
			execPath = "."
		} else {
			execPath = filepath.Dir(execPath)
		}

		// 在可执行文件目录和当前目录中搜索配置文件
		viper.AddConfigPath(execPath)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".ora2pg-admin")
	}

	viper.AutomaticEnv() // 读取匹配的环境变量

	// 如果找到配置文件，则读取它
	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			logrus.Infof("使用配置文件: %s", viper.ConfigFileUsed())
		}
	}
}

// initLogger 初始化日志系统
func initLogger() {
	// 创建日志配置
	logConfig := &utils.LogConfig{
		Format:     "text",
		Output:     "stdout",
		TimeFormat: "2006-01-02 15:04:05",
	}

	// 根据参数设置日志级别
	if quiet {
		logConfig.Level = utils.LogLevelError
	} else if verbose {
		logConfig.Level = utils.LogLevelDebug
	} else {
		logConfig.Level = utils.LogLevelInfo
	}

	// 如果指定了日志文件，则输出到文件
	if logFile != "" {
		logConfig.Output = "file"
		logConfig.FilePath = logFile
	}

	// 初始化全局日志器
	utils.InitGlobalLogger(logConfig)
}
