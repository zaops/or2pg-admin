package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"ora2pg-admin/internal/oracle"
)

// statusCmd 显示当前项目状态
var statusCmd = &cobra.Command{
	Use:   "状态",
	Short: "查看当前项目状态",
	Long:  "显示当前迁移项目的状态信息，包括配置文件、环境检查结果等。",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📊 当前项目状态")
		fmt.Println()

		// 检查配置文件
		configFile := viper.ConfigFileUsed()
		if configFile != "" {
			fmt.Printf("✅ 配置文件: %s\n", configFile)
		} else {
			fmt.Println("❌ 未找到配置文件")
		}

		// 检查项目目录
		if checkProjectDirectory() {
			fmt.Println("✅ 项目目录: 已初始化")
		} else {
			fmt.Println("❌ 项目目录: 未初始化")
		}

		// 检查ora2pg二进制
		if checkOra2pgBinary() {
			fmt.Println("✅ ora2pg: 已安装")
		} else {
			fmt.Println("❌ ora2pg: 未找到")
		}

		// 检查Oracle客户端
		detector := oracle.NewClientDetector()
		clientInfo, err := detector.DetectClient()
		if err != nil {
			fmt.Printf("❌ Oracle客户端: 检测失败 (%v)\n", err)
		} else if clientInfo.Installed {
			if clientInfo.Version != "" {
				fmt.Printf("✅ Oracle客户端: %s\n", clientInfo.Version)
			} else {
				fmt.Println("✅ Oracle客户端: 已安装")
			}
		} else {
			fmt.Println("❌ Oracle客户端: 未安装")
		}

		// 显示当前工作目录
		wd, err := os.Getwd()
		if err != nil {
			logrus.Warnf("无法获取当前工作目录: %v", err)
		} else {
			fmt.Printf("📁 工作目录: %s\n", wd)
		}

		fmt.Println()
		fmt.Println("💡 提示: 使用 'ora2pg-admin 帮助' 查看可用命令")
	},
}

// checkProjectDirectory 检查项目目录是否已初始化
func checkProjectDirectory() bool {
	// 检查是否存在 .ora2pg-admin 目录
	if _, err := os.Stat(".ora2pg-admin"); err == nil {
		return true
	}
	return false
}

// checkOra2pgBinary 检查ora2pg二进制文件是否可用
func checkOra2pgBinary() bool {
	// 在PATH中查找ora2pg
	_, err := exec.LookPath("ora2pg")
	return err == nil
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
