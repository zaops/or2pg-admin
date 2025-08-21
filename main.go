package main

import (
	"fmt"
	"os"

	"ora2pg-admin/cmd"
)

// 版本信息
var (
	Version   = "1.0.0"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// 设置版本信息
	cmd.SetVersionInfo(Version, BuildTime, GitCommit)
	
	// 执行根命令
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "执行命令时发生错误: %v\n", err)
		os.Exit(1)
	}
}
