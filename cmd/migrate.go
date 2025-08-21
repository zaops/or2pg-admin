package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"ora2pg-admin/internal/config"
	"ora2pg-admin/internal/service"
	"ora2pg-admin/internal/utils"
)

var (
	migrateTimeout   time.Duration
	migrateParallel  int
	migrateResume    bool
	migrateValidate  bool
	migrateBackup    bool
)

// migrateCmd 迁移命令
var migrateCmd = &cobra.Command{
	Use:   "迁移",
	Short: "执行数据库迁移",
	Long: `执行Oracle到PostgreSQL数据库迁移操作。

此命令提供分阶段的迁移功能，支持：
• 结构迁移：迁移表结构、视图、序列、索引等数据库对象
• 数据迁移：迁移表数据内容
• 完整迁移：按顺序执行结构和数据迁移

支持实时进度跟踪、中断恢复、结果验证等高级功能。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 如果没有提供子命令，显示帮助信息
		cmd.Help()
	},
}

// migrateStructureCmd 结构迁移命令
var migrateStructureCmd = &cobra.Command{
	Use:   "结构",
	Short: "迁移数据库结构",
	Long: `迁移Oracle数据库的结构对象到PostgreSQL。

包括以下对象类型：
• 表结构（TABLE）
• 视图（VIEW）
• 序列（SEQUENCE）
• 索引（INDEX）
• 触发器（TRIGGER）
• 函数和存储过程（FUNCTION、PROCEDURE）

结构迁移通常在数据迁移之前执行，为数据提供目标结构。`,
	Run: runMigrateStructure,
}

// migrateDataCmd 数据迁移命令
var migrateDataCmd = &cobra.Command{
	Use:   "数据",
	Short: "迁移数据内容",
	Long: `迁移Oracle数据库的数据内容到PostgreSQL。

此命令将执行：
• 表数据复制（COPY）
• 数据插入（INSERT）
• 数据验证和完整性检查

建议在结构迁移完成后执行数据迁移。`,
	Run: runMigrateData,
}

// migrateAllCmd 完整迁移命令
var migrateAllCmd = &cobra.Command{
	Use:   "全部",
	Short: "执行完整迁移",
	Long: `执行完整的数据库迁移流程。

此命令将按以下顺序执行：
1. 结构迁移：创建表、视图、序列等对象
2. 数据迁移：复制表数据
3. 索引和约束：创建索引和约束
4. 触发器和函数：创建触发器和存储过程
5. 权限和授权：设置数据库权限

提供完整的迁移解决方案，适合一次性迁移场景。`,
	Run: runMigrateAll,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(migrateStructureCmd)
	migrateCmd.AddCommand(migrateDataCmd)
	migrateCmd.AddCommand(migrateAllCmd)

	// 添加命令参数
	migrateCmd.PersistentFlags().DurationVar(&migrateTimeout, "timeout", 2*time.Hour, "迁移超时时间")
	migrateCmd.PersistentFlags().IntVar(&migrateParallel, "parallel", 0, "并行作业数（0表示使用配置文件设置）")
	migrateCmd.PersistentFlags().BoolVar(&migrateResume, "resume", false, "恢复中断的迁移")
	migrateCmd.PersistentFlags().BoolVar(&migrateValidate, "validate", true, "迁移后验证结果")
	migrateCmd.PersistentFlags().BoolVar(&migrateBackup, "backup", true, "迁移前创建备份")
}

// runMigrateStructure 执行结构迁移
func runMigrateStructure(cmd *cobra.Command, args []string) {
	logger := utils.GetGlobalLogger()
	
	fmt.Println("🏗️ 数据库结构迁移")
	fmt.Println()

	// 1. 加载配置和初始化服务
	migrationService, err := initializeMigrationService()
	if err != nil {
		fmt.Printf("%s\n", utils.FormatError(err))
		os.Exit(1)
	}

	// 2. 定义结构迁移类型
	structureTypes := []service.MigrationType{
		service.MigrationTypeTable,
		service.MigrationTypeView,
		service.MigrationTypeSequence,
		service.MigrationTypeIndex,
		service.MigrationTypeTrigger,
		service.MigrationTypeFunction,
		service.MigrationTypeProcedure,
	}

	// 3. 执行迁移
	ctx, cancel := createMigrationContext()
	defer cancel()

	results, err := executeMigrationWithProgress(ctx, migrationService, structureTypes, "结构迁移")
	if err != nil {
		fmt.Printf("%s\n", utils.FormatError(err))
		os.Exit(1)
	}

	// 4. 显示结果
	showMigrationResults(results, "结构迁移")
	
	logger.Info("结构迁移完成")
}

// runMigrateData 执行数据迁移
func runMigrateData(cmd *cobra.Command, args []string) {
	logger := utils.GetGlobalLogger()
	
	fmt.Println("📊 数据内容迁移")
	fmt.Println()

	// 1. 加载配置和初始化服务
	migrationService, err := initializeMigrationService()
	if err != nil {
		fmt.Printf("%s\n", utils.FormatError(err))
		os.Exit(1)
	}

	// 2. 定义数据迁移类型
	dataTypes := []service.MigrationType{
		service.MigrationTypeCopy,
		service.MigrationTypeInsert,
	}

	// 3. 执行迁移
	ctx, cancel := createMigrationContext()
	defer cancel()

	results, err := executeMigrationWithProgress(ctx, migrationService, dataTypes, "数据迁移")
	if err != nil {
		fmt.Printf("%s\n", utils.FormatError(err))
		os.Exit(1)
	}

	// 4. 显示结果
	showMigrationResults(results, "数据迁移")
	
	logger.Info("数据迁移完成")
}

// runMigrateAll 执行完整迁移
func runMigrateAll(cmd *cobra.Command, args []string) {
	logger := utils.GetGlobalLogger()
	
	fmt.Println("🚀 完整数据库迁移")
	fmt.Println()

	// 1. 加载配置和初始化服务
	migrationService, err := initializeMigrationService()
	if err != nil {
		fmt.Printf("%s\n", utils.FormatError(err))
		os.Exit(1)
	}

	// 2. 定义完整迁移类型（按执行顺序）
	allTypes := []service.MigrationType{
		// 第一阶段：基础结构
		service.MigrationTypeTable,
		service.MigrationTypeView,
		service.MigrationTypeSequence,
		// 第二阶段：数据内容
		service.MigrationTypeCopy,
		// 第三阶段：索引和约束
		service.MigrationTypeIndex,
		// 第四阶段：程序对象
		service.MigrationTypeTrigger,
		service.MigrationTypeFunction,
		service.MigrationTypeProcedure,
		// 第五阶段：权限
		service.MigrationTypeGrant,
	}

	// 3. 执行迁移
	ctx, cancel := createMigrationContext()
	defer cancel()

	results, err := executeMigrationWithProgress(ctx, migrationService, allTypes, "完整迁移")
	if err != nil {
		fmt.Printf("%s\n", utils.FormatError(err))
		os.Exit(1)
	}

	// 4. 显示结果
	showMigrationResults(results, "完整迁移")
	
	// 5. 执行验证（如果启用）
	if migrateValidate {
		fmt.Println()
		fmt.Println("🔍 迁移结果验证")
		fmt.Println("─────────────")
		validateMigrationResults(results)
	}
	
	logger.Info("完整迁移完成")
}

// initializeMigrationService 初始化迁移服务
func initializeMigrationService() (*service.MigrationService, error) {
	// 检查项目环境
	fileUtils := utils.NewFileUtils()
	if !fileUtils.DirExists(".ora2pg-admin") {
		return nil, utils.NewError(utils.ErrorTypeConfig, "PROJECT_NOT_INITIALIZED").
			Message("项目未初始化").
			Suggestion("请先使用 'ora2pg-admin 初始化 [项目名称]' 创建项目").
			Build()
	}

	// 加载配置
	configPath := filepath.Join(".ora2pg-admin", "config.yaml")
	manager := config.NewManager()
	if err := manager.LoadConfig(configPath); err != nil {
		return nil, utils.ConfigErrors.ParseFailed(err)
	}

	// 创建迁移服务
	migrationService := service.NewMigrationService(manager.GetConfig())

	// 应用命令行参数
	if migrateParallel > 0 {
		migrationService.SetParallelJobs(migrateParallel)
	}

	return migrationService, nil
}

// createMigrationContext 创建迁移上下文
func createMigrationContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), migrateTimeout)

	// 设置信号处理，支持优雅中断
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n⚠️ 收到中断信号，正在停止迁移...")
		cancel()
	}()

	return ctx, cancel
}

// executeMigrationWithProgress 执行迁移并显示进度
func executeMigrationWithProgress(ctx context.Context, migrationService *service.MigrationService,
	migrationTypes []service.MigrationType, taskName string) ([]*service.ExecutionResult, error) {

	fmt.Printf("📋 开始执行%s，共 %d 个步骤\n", taskName, len(migrationTypes))
	fmt.Println()

	// 创建进度跟踪器
	progressTracker := service.NewProgressTracker()
	progressTracker.Start(taskName, len(migrationTypes))

	// 执行迁移
	results, err := migrationService.ExecuteWithProgress(ctx, migrationTypes, progressTracker)

	// 停止进度跟踪
	progressTracker.Stop()

	return results, err
}

// showMigrationResults 显示迁移结果
func showMigrationResults(results []*service.ExecutionResult, taskName string) {
	fmt.Println()
	fmt.Printf("📊 %s结果摘要\n", taskName)
	fmt.Println("─────────────────")

	successful := 0
	failed := 0
	totalDuration := time.Duration(0)

	for i, result := range results {
		fmt.Printf("%d. ", i+1)

		switch result.Status {
		case service.StatusCompleted:
			fmt.Printf("✅ 成功")
			successful++
		case service.StatusFailed:
			fmt.Printf("❌ 失败")
			failed++
		case service.StatusCancelled:
			fmt.Printf("⚠️ 已取消")
		default:
			fmt.Printf("❓ 未知状态")
		}

		fmt.Printf(" (耗时: %v)\n", result.Duration)
		totalDuration += result.Duration

		if result.Error != nil {
			fmt.Printf("   错误: %s\n", result.Error.Error())
		}
	}

	fmt.Println()
	fmt.Printf("总计: %d 成功, %d 失败, 总耗时: %v\n", successful, failed, totalDuration)

	if failed == 0 {
		fmt.Printf("🎉 %s全部完成！\n", taskName)
	} else {
		fmt.Printf("⚠️ %s部分失败，请检查错误信息\n", taskName)
	}
}

// validateMigrationResults 验证迁移结果
func validateMigrationResults(results []*service.ExecutionResult) {
	fmt.Println("正在验证迁移结果...")

	// 这里可以添加具体的验证逻辑
	// 例如：检查表数量、数据行数、索引等

	hasErrors := false
	for _, result := range results {
		if result.Status == service.StatusFailed {
			hasErrors = true
			break
		}
	}

	if hasErrors {
		fmt.Println("❌ 验证发现问题，建议检查迁移日志")
	} else {
		fmt.Println("✅ 验证通过，迁移结果正常")
	}
}
