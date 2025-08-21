package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
	"ora2pg-admin/internal/config"
	"ora2pg-admin/internal/utils"
)

// configureMigrationTypes 配置迁移类型
func configureMigrationTypes(migrationConfig *config.MigrationConfig) error {
	// 可用的迁移类型
	availableTypes := []string{
		"TABLE - 表结构和数据",
		"VIEW - 视图",
		"SEQUENCE - 序列",
		"INDEX - 索引",
		"TRIGGER - 触发器",
		"FUNCTION - 函数",
		"PROCEDURE - 存储过程",
		"PACKAGE - 包",
		"TYPE - 自定义类型",
		"GRANT - 权限",
		"TABLESPACE - 表空间",
		"PARTITION - 分区",
	}

	typeMap := map[string]string{
		"TABLE - 表结构和数据":    "TABLE",
		"VIEW - 视图":        "VIEW",
		"SEQUENCE - 序列":    "SEQUENCE",
		"INDEX - 索引":       "INDEX",
		"TRIGGER - 触发器":    "TRIGGER",
		"FUNCTION - 函数":    "FUNCTION",
		"PROCEDURE - 存储过程": "PROCEDURE",
		"PACKAGE - 包":       "PACKAGE",
		"TYPE - 自定义类型":     "TYPE",
		"GRANT - 权限":       "GRANT",
		"TABLESPACE - 表空间": "TABLESPACE",
		"PARTITION - 分区":   "PARTITION",
	}

	// 显示当前配置
	if len(migrationConfig.Types) > 0 {
		fmt.Printf("当前迁移类型: %s\n", strings.Join(migrationConfig.Types, ", "))
		fmt.Println()
	}

	// 多选提示
	fmt.Println("请选择要迁移的对象类型（使用空格选择/取消选择，回车确认）:")
	
	// 创建选择器
	prompt := promptui.Select{
		Label: "迁移类型选择",
		Items: availableTypes,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}?",
			Active:   "▶ {{ . | cyan }}",
			Inactive: "  {{ . | white }}",
			Selected: "✓ {{ . | green }}",
		},
	}

	selectedTypes := make(map[string]bool)
	
	// 预选择当前配置的类型
	for _, currentType := range migrationConfig.Types {
		for display, value := range typeMap {
			if value == currentType {
				selectedTypes[display] = true
				break
			}
		}
	}

	// 简化的多选实现
	fmt.Println("请逐个选择要迁移的类型（选择 'DONE' 完成选择）:")
	
	// 添加完成选项
	selectionItems := append(availableTypes, "DONE - 完成选择")
	
	for {
		prompt.Items = selectionItems
		
		// 显示当前选择状态
		fmt.Println("\n当前已选择:")
		hasSelection := false
		for item, selected := range selectedTypes {
			if selected {
				fmt.Printf("  ✓ %s\n", item)
				hasSelection = true
			}
		}
		if !hasSelection {
			fmt.Println("  (无)")
		}
		fmt.Println()

		_, result, err := prompt.Run()
		if err != nil {
			return utils.NewError(utils.ErrorTypeUser, "INPUT_CANCELLED").
				Message("用户取消了选择").Build()
		}

		if result == "DONE - 完成选择" {
			break
		}

		// 切换选择状态
		selectedTypes[result] = !selectedTypes[result]
	}

	// 转换为配置格式
	var newTypes []string
	for display, selected := range selectedTypes {
		if selected {
			if typeValue, exists := typeMap[display]; exists {
				newTypes = append(newTypes, typeValue)
			}
		}
	}

	if len(newTypes) == 0 {
		return utils.ValidationErrors.Required("迁移类型")
	}

	migrationConfig.Types = newTypes
	fmt.Printf("✅ 已选择 %d 种迁移类型\n", len(newTypes))
	return nil
}

// configurePerformanceSettings 配置性能参数
func configurePerformanceSettings(migrationConfig *config.MigrationConfig) error {
	// 配置并行作业数
	jobsPrompt := promptui.Prompt{
		Label:    "并行作业数（建议设置为CPU核心数）",
		Default:  strconv.Itoa(migrationConfig.ParallelJobs),
		Validate: validatePositiveInt,
	}
	jobsStr, err := jobsPrompt.Run()
	if err != nil {
		return utils.NewError(utils.ErrorTypeUser, "INPUT_CANCELLED").
			Message("用户取消了输入").Build()
	}
	if jobs, err := strconv.Atoi(strings.TrimSpace(jobsStr)); err == nil {
		migrationConfig.ParallelJobs = jobs
	}

	// 配置批处理大小
	batchPrompt := promptui.Prompt{
		Label:    "批处理大小（每次处理的记录数）",
		Default:  strconv.Itoa(migrationConfig.BatchSize),
		Validate: validatePositiveInt,
	}
	batchStr, err := batchPrompt.Run()
	if err != nil {
		return utils.NewError(utils.ErrorTypeUser, "INPUT_CANCELLED").
			Message("用户取消了输入").Build()
	}
	if batch, err := strconv.Atoi(strings.TrimSpace(batchStr)); err == nil {
		migrationConfig.BatchSize = batch
	}

	// 配置输出目录
	outputPrompt := promptui.Prompt{
		Label:   "输出目录",
		Default: migrationConfig.OutputDir,
		Validate: validateRequired,
	}
	output, err := outputPrompt.Run()
	if err != nil {
		return utils.NewError(utils.ErrorTypeUser, "INPUT_CANCELLED").
			Message("用户取消了输入").Build()
	}
	migrationConfig.OutputDir = strings.TrimSpace(output)

	fmt.Println("✅ 性能参数配置完成")
	return nil
}

// configureAdvancedOptions 配置高级选项
func configureAdvancedOptions(migrationConfig *config.MigrationConfig) error {
	// 配置日志级别
	logLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	logPrompt := promptui.Select{
		Label: "选择日志级别",
		Items: logLevels,
	}
	
	// 找到当前日志级别的索引
	currentIndex := 1 // 默认INFO
	for i, level := range logLevels {
		if level == migrationConfig.LogLevel {
			currentIndex = i
			break
		}
	}
	logPrompt.CursorPos = currentIndex

	_, logLevel, err := logPrompt.Run()
	if err != nil {
		return utils.NewError(utils.ErrorTypeUser, "INPUT_CANCELLED").
			Message("用户取消了选择").Build()
	}
	migrationConfig.LogLevel = logLevel

	fmt.Println("✅ 高级选项配置完成")
	return nil
}

// previewMigrationConfig 预览迁移配置
func previewMigrationConfig(migrationConfig *config.MigrationConfig) {
	fmt.Printf("迁移类型: %s\n", strings.Join(migrationConfig.Types, ", "))
	fmt.Printf("并行作业数: %d\n", migrationConfig.ParallelJobs)
	fmt.Printf("批处理大小: %d\n", migrationConfig.BatchSize)
	fmt.Printf("输出目录: %s\n", migrationConfig.OutputDir)
	fmt.Printf("日志级别: %s\n", migrationConfig.LogLevel)
}

// confirmConfiguration 确认配置
func confirmConfiguration() bool {
	prompt := promptui.Prompt{
		Label:     "确认保存配置",
		IsConfirm: true,
	}
	
	_, err := prompt.Run()
	return err == nil
}

// generateOra2pgConfig 生成ora2pg配置文件
func generateOra2pgConfig(cfg *config.ProjectConfig) error {
	templateEngine := config.NewTemplateEngine("templates")
	
	// 检查模板目录是否存在
	fileUtils := utils.NewFileUtils()
	if !fileUtils.DirExists("templates") {
		// 尝试使用可执行文件目录的模板
		if execPath, err := fileUtils.GetExecutablePath(); err == nil {
			templateDir := fileUtils.JoinPath(execPath, "templates")
			if fileUtils.DirExists(templateDir) {
				templateEngine.SetTemplateDir(templateDir)
			} else {
				return utils.NewError(utils.ErrorTypeFile, "TEMPLATE_NOT_FOUND").
					Message("未找到ora2pg配置模板").
					Suggestion("请确认templates目录存在").
					Build()
			}
		}
	}

	// 生成ora2pg配置文件
	outputPath := fileUtils.JoinPath(cfg.Migration.OutputDir, "ora2pg.conf")
	if err := templateEngine.GenerateOra2pgConfig(cfg, outputPath); err != nil {
		return err
	}

	fmt.Printf("✅ 已生成ora2pg配置文件: %s\n", outputPath)
	return nil
}

// validatePositiveInt 验证正整数
func validatePositiveInt(input string) error {
	input = strings.TrimSpace(input)
	if input == "" {
		return fmt.Errorf("不能为空")
	}
	
	value, err := strconv.Atoi(input)
	if err != nil {
		return fmt.Errorf("必须是数字")
	}
	
	if value <= 0 {
		return fmt.Errorf("必须是正整数")
	}
	
	return nil
}
