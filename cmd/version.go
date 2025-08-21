package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd 显示版本信息
var versionCmd = &cobra.Command{
	Use:   "版本",
	Short: "显示版本信息",
	Long:  "显示 ora2pg-admin 的版本信息，包括版本号、构建时间和Git提交哈希。",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ora2pg-admin 版本信息:\n")
		fmt.Printf("  版本号:   %s\n", version)
		fmt.Printf("  构建时间: %s\n", buildTime)
		fmt.Printf("  Git提交:  %s\n", gitCommit)
		fmt.Printf("\n")
		fmt.Printf("一个友好的 ora2pg 中文命令行管理工具\n")
		fmt.Printf("用于简化 Oracle 到 PostgreSQL 数据库迁移操作\n")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
