package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// helpCmd 显示帮助信息
var helpCmd = &cobra.Command{
	Use:   "帮助",
	Short: "显示帮助信息",
	Long:  "显示 ora2pg-admin 的详细帮助信息和使用指南。",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚀 Ora2Pg 中文CLI管理器 - 使用指南")
		fmt.Println()
		fmt.Println("📋 主要命令:")
		fmt.Println("  初始化 [项目名]     创建新的迁移项目")
		fmt.Println("  配置 数据库         配置Oracle和PostgreSQL连接")
		fmt.Println("  配置 选项           配置迁移选项和参数")
		fmt.Println("  检查 环境           检查Oracle客户端等环境")
		fmt.Println("  检查 连接           测试数据库连接")
		fmt.Println("  迁移 结构           迁移数据库结构")
		fmt.Println("  迁移 数据           迁移数据内容")
		fmt.Println("  迁移 全部           完整迁移流程")
		fmt.Println("  状态               查看当前项目状态")
		fmt.Println("  版本               显示版本信息")
		fmt.Println("  帮助               显示此帮助信息")
		fmt.Println()
		fmt.Println("🔧 全局参数:")
		fmt.Println("  --config, -c       指定配置文件路径")
		fmt.Println("  --verbose, -v      显示详细输出")
		fmt.Println("  --quiet, -q        静默模式")
		fmt.Println("  --dry-run          预览模式，不执行实际操作")
		fmt.Println("  --log-file         指定日志文件路径")
		fmt.Println()
		fmt.Println("💡 典型使用流程:")
		fmt.Println("  1. ora2pg-admin 初始化 我的迁移项目")
		fmt.Println("  2. ora2pg-admin 检查 环境")
		fmt.Println("  3. ora2pg-admin 配置 数据库")
		fmt.Println("  4. ora2pg-admin 检查 连接")
		fmt.Println("  5. ora2pg-admin 迁移 全部")
		fmt.Println()
		fmt.Println("📚 更多信息请查看项目文档或访问 GitHub 仓库")
	},
}

func init() {
	rootCmd.AddCommand(helpCmd)
}
