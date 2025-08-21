package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"ora2pg-admin/internal/config"
	"ora2pg-admin/internal/utils"
)

var (
	initForce       bool
	initTemplate    string
	initDescription string
)

// initCmd 初始化命令
var initCmd = &cobra.Command{
	Use:   "初始化 [项目名称]",
	Short: "创建新的迁移项目",
	Long: `创建新的Oracle到PostgreSQL数据库迁移项目。

此命令将创建项目目录结构，生成基础配置文件，并提供项目模板选择。
项目初始化后，您可以使用其他命令进行配置和迁移操作。

示例:
  ora2pg-admin 初始化 我的迁移项目
  ora2pg-admin 初始化 --template=basic --description="生产环境迁移" 生产迁移`,
	Args: cobra.MaximumNArgs(1),
	Run:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	// 添加命令参数
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "强制覆盖已存在的项目")
	initCmd.Flags().StringVarP(&initTemplate, "template", "t", "", "项目模板 (basic, advanced, custom)")
	initCmd.Flags().StringVarP(&initDescription, "description", "d", "", "项目描述")
}

// runInit 执行初始化命令
func runInit(cmd *cobra.Command, args []string) {
	logger := utils.GetGlobalLogger()
	fileUtils := utils.NewFileUtils()

	fmt.Println("🚀 Ora2Pg 项目初始化向导")
	fmt.Println()

	// 1. 获取项目名称
	projectName, err := getProjectName(args)
	if err != nil {
		fmt.Printf("%s\n", utils.FormatError(err))
		os.Exit(1)
	}

	// 2. 检查项目是否已存在
	if err := checkProjectExists(projectName, fileUtils); err != nil {
		fmt.Printf("%s\n", utils.FormatError(err))
		os.Exit(1)
	}

	// 3. 收集项目信息
	projectInfo, err := collectProjectInfo(projectName)
	if err != nil {
		fmt.Printf("%s\n", utils.FormatError(err))
		os.Exit(1)
	}

	// 4. 创建项目目录结构
	fmt.Println("📁 创建项目目录结构...")
	if err := createProjectStructure(projectName, fileUtils); err != nil {
		fmt.Printf("%s\n", utils.FormatError(err))
		os.Exit(1)
	}

	// 5. 生成配置文件
	fmt.Println("⚙️ 生成项目配置文件...")
	if err := generateProjectConfig(projectName, projectInfo, fileUtils); err != nil {
		fmt.Printf("%s\n", utils.FormatError(err))
		os.Exit(1)
	}

	// 6. 创建示例文件
	fmt.Println("📄 创建示例文件...")
	if err := createExampleFiles(projectName, fileUtils); err != nil {
		logger.Warnf("创建示例文件时出现警告: %v", err)
	}

	// 7. 显示成功信息和后续指导
	showSuccessMessage(projectName, projectInfo)
}

// getProjectName 获取项目名称
func getProjectName(args []string) (string, error) {
	if len(args) > 0 {
		return strings.TrimSpace(args[0]), nil
	}

	// 如果没有提供项目名称，通过交互式输入获取
	prompt := promptui.Prompt{
		Label:    "请输入项目名称",
		Default:  "我的迁移项目",
		Validate: validateProjectName,
	}

	result, err := prompt.Run()
	if err != nil {
		return "", utils.NewError(utils.ErrorTypeUser, "INPUT_CANCELLED").
			Message("用户取消了输入").
			Build()
	}

	return strings.TrimSpace(result), nil
}

// validateProjectName 验证项目名称
func validateProjectName(input string) error {
	input = strings.TrimSpace(input)
	if input == "" {
		return fmt.Errorf("项目名称不能为空")
	}
	if len(input) > 50 {
		return fmt.Errorf("项目名称长度不能超过50个字符")
	}
	// 检查是否包含非法字符
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(input, char) {
			return fmt.Errorf("项目名称不能包含字符: %s", char)
		}
	}
	return nil
}

// checkProjectExists 检查项目是否已存在
func checkProjectExists(projectName string, fileUtils *utils.FileUtils) error {
	projectDir := getProjectDir(projectName)
	
	if fileUtils.DirExists(projectDir) {
		if !initForce {
			return utils.NewError(utils.ErrorTypeUser, "PROJECT_EXISTS").
				Message(fmt.Sprintf("项目目录已存在: %s", projectDir)).
				Suggestion("使用 --force 参数强制覆盖已存在的项目").
				Suggestion("或者选择不同的项目名称").
				Build()
		}
		
		// 如果使用了 --force 参数，询问确认
		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("项目目录 %s 已存在，是否覆盖", projectDir),
			IsConfirm: true,
		}
		
		_, err := prompt.Run()
		if err != nil {
			return utils.NewError(utils.ErrorTypeUser, "OPERATION_CANCELLED").
				Message("用户取消了覆盖操作").
				Build()
		}
		
		// 删除已存在的项目目录
		if err := fileUtils.RemoveDir(projectDir); err != nil {
			return utils.FileErrors.CreateFailed(projectDir, err)
		}
	}
	
	return nil
}

// ProjectInfo 项目信息
type ProjectInfo struct {
	Name        string
	Description string
	Template    string
	Author      string
	Email       string
}

// collectProjectInfo 收集项目信息
func collectProjectInfo(projectName string) (*ProjectInfo, error) {
	info := &ProjectInfo{
		Name: projectName,
	}

	// 获取项目描述
	if initDescription != "" {
		info.Description = initDescription
	} else {
		prompt := promptui.Prompt{
			Label:   "请输入项目描述（可选）",
			Default: "Oracle到PostgreSQL数据库迁移项目",
		}
		result, err := prompt.Run()
		if err == nil {
			info.Description = strings.TrimSpace(result)
		}
	}

	// 获取项目模板
	if initTemplate != "" {
		info.Template = initTemplate
	} else {
		templates := []string{"basic", "advanced", "custom"}
		prompt := promptui.Select{
			Label: "选择项目模板",
			Items: []string{
				"basic - 基础模板（推荐新手使用）",
				"advanced - 高级模板（包含更多配置选项）",
				"custom - 自定义模板（手动配置所有选项）",
			},
		}
		
		index, _, err := prompt.Run()
		if err != nil {
			info.Template = "basic" // 默认使用基础模板
		} else {
			info.Template = templates[index]
		}
	}

	// 获取作者信息（可选）- 仅在交互模式下询问
	if initDescription == "" && initTemplate == "" {
		prompt := promptui.Prompt{
			Label:   "请输入作者姓名（可选，直接回车跳过）",
			Default: "",
		}
		if result, err := prompt.Run(); err == nil && strings.TrimSpace(result) != "" {
			info.Author = strings.TrimSpace(result)

			// 获取邮箱信息（可选）
			emailPrompt := promptui.Prompt{
				Label:   "请输入邮箱地址（可选，直接回车跳过）",
				Default: "",
			}
			if emailResult, emailErr := emailPrompt.Run(); emailErr == nil {
				info.Email = strings.TrimSpace(emailResult)
			}
		}
	}

	return info, nil
}

// getProjectDir 获取项目目录路径
func getProjectDir(projectName string) string {
	// 将项目名称转换为合法的目录名
	dirName := strings.ReplaceAll(projectName, " ", "_")
	dirName = strings.ToLower(dirName)
	return dirName
}

// createProjectStructure 创建项目目录结构
func createProjectStructure(projectName string, fileUtils *utils.FileUtils) error {
	projectDir := getProjectDir(projectName)
	
	// 需要创建的目录列表
	directories := []string{
		projectDir,
		filepath.Join(projectDir, ".ora2pg-admin"),
		filepath.Join(projectDir, "logs"),
		filepath.Join(projectDir, "output"),
		filepath.Join(projectDir, "scripts"),
		filepath.Join(projectDir, "backup"),
		filepath.Join(projectDir, "docs"),
	}

	// 创建目录
	for _, dir := range directories {
		if err := fileUtils.EnsureDir(dir); err != nil {
			return utils.FileErrors.CreateFailed(dir, err)
		}
		fmt.Printf("  ✅ %s\n", dir)
	}

	return nil
}

// generateProjectConfig 生成项目配置文件
func generateProjectConfig(projectName string, projectInfo *ProjectInfo, fileUtils *utils.FileUtils) error {
	projectDir := getProjectDir(projectName)
	configPath := filepath.Join(projectDir, ".ora2pg-admin", "config.yaml")

	// 创建配置管理器
	manager := config.NewManager()
	manager.CreateDefaultConfig(projectName)

	// 获取配置并更新项目信息
	cfg := manager.GetConfig()
	cfg.Project.Name = projectInfo.Name
	cfg.Project.Description = projectInfo.Description
	cfg.Project.Created = time.Now()
	cfg.Project.Updated = time.Now()

	// 根据模板调整配置
	switch projectInfo.Template {
	case "basic":
		// 基础模板：简化配置
		cfg.Migration.Types = []string{"TABLE", "VIEW", "SEQUENCE"}
		cfg.Migration.ParallelJobs = 2
	case "advanced":
		// 高级模板：完整配置
		cfg.Migration.Types = []string{"TABLE", "VIEW", "SEQUENCE", "INDEX", "TRIGGER", "FUNCTION"}
		cfg.Migration.ParallelJobs = 4
	case "custom":
		// 自定义模板：保持默认配置，用户后续自行配置
	}

	// 保存配置文件
	if err := manager.SaveConfig(configPath); err != nil {
		return utils.ConfigErrors.ParseFailed(err)
	}

	fmt.Printf("  ✅ %s\n", configPath)
	return nil
}

// createExampleFiles 创建示例文件
func createExampleFiles(projectName string, fileUtils *utils.FileUtils) error {
	projectDir := getProjectDir(projectName)

	// 创建README文件
	readmePath := filepath.Join(projectDir, "README.md")
	readmeContent := generateReadmeContent(projectName)
	if err := fileUtils.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return err
	}
	fmt.Printf("  ✅ %s\n", readmePath)

	// 创建.gitignore文件
	gitignorePath := filepath.Join(projectDir, ".gitignore")
	gitignoreContent := generateGitignoreContent()
	if err := fileUtils.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		return err
	}
	fmt.Printf("  ✅ %s\n", gitignorePath)

	// 创建示例脚本
	scriptPath := filepath.Join(projectDir, "scripts", "example.sql")
	scriptContent := generateExampleScript()
	if err := fileUtils.WriteFile(scriptPath, []byte(scriptContent), 0644); err != nil {
		return err
	}
	fmt.Printf("  ✅ %s\n", scriptPath)

	return nil
}

// generateReadmeContent 生成README内容
func generateReadmeContent(projectName string) string {
	return fmt.Sprintf(`# %s

这是一个使用 ora2pg-admin 创建的Oracle到PostgreSQL数据库迁移项目。

## 项目结构

- .ora2pg-admin/ - 项目配置文件
- logs/ - 日志文件目录
- output/ - 迁移输出文件目录
- scripts/ - 自定义SQL脚本目录
- backup/ - 备份文件目录
- docs/ - 项目文档目录

## 快速开始

1. 配置数据库连接：
   `+"`"+`bash
   ora2pg-admin 配置 数据库
   `+"`"+`

2. 检查环境：
   `+"`"+`bash
   ora2pg-admin 检查 环境
   `+"`"+`

3. 测试连接：
   `+"`"+`bash
   ora2pg-admin 检查 连接
   `+"`"+`

4. 执行迁移：
   `+"`"+`bash
   ora2pg-admin 迁移 全部
   `+"`"+`

## 配置文件

主要配置文件位于 .ora2pg-admin/config.yaml，包含：
- Oracle数据库连接配置
- PostgreSQL数据库连接配置
- 迁移选项和参数设置

## 注意事项

- 请确保已安装Oracle客户端
- 建议在迁移前进行数据备份
- 大型数据库迁移建议分批进行

## 帮助

使用以下命令获取帮助：
`+"`"+`bash
ora2pg-admin 帮助
`+"`"+`

---
*此项目由 ora2pg-admin 自动生成*
`, projectName)
}

// generateGitignoreContent 生成.gitignore内容
func generateGitignoreContent() string {
	return `# 日志文件
logs/
*.log

# 输出文件
output/
*.sql
*.dump

# 备份文件
backup/
*.bak

# 临时文件
*.tmp
*.temp

# 敏感配置文件
config.local.yaml
.env
.env.local

# 操作系统文件
.DS_Store
Thumbs.db

# IDE文件
.vscode/
.idea/
*.swp
*.swo
`
}

// generateExampleScript 生成示例脚本
func generateExampleScript() string {
	return `-- 示例SQL脚本
-- 此文件可以包含自定义的SQL语句，用于迁移前后的数据处理

-- 示例：创建索引
-- CREATE INDEX idx_example ON table_name (column_name);

-- 示例：数据清理
-- DELETE FROM table_name WHERE condition;

-- 示例：数据转换
-- UPDATE table_name SET column_name = REPLACE(column_name, 'old_value', 'new_value');
`
}

// showSuccessMessage 显示成功信息和后续指导
func showSuccessMessage(projectName string, projectInfo *ProjectInfo) {
	projectDir := getProjectDir(projectName)

	fmt.Println()
	fmt.Println("🎉 项目初始化成功！")
	fmt.Println()
	fmt.Printf("📁 项目目录: %s\n", projectDir)
	fmt.Printf("📋 项目名称: %s\n", projectInfo.Name)
	fmt.Printf("📝 项目描述: %s\n", projectInfo.Description)
	fmt.Printf("🎨 项目模板: %s\n", projectInfo.Template)
	if projectInfo.Author != "" {
		fmt.Printf("👤 作者: %s\n", projectInfo.Author)
	}
	if projectInfo.Email != "" {
		fmt.Printf("📧 邮箱: %s\n", projectInfo.Email)
	}

	fmt.Println()
	fmt.Println("🚀 后续步骤:")
	fmt.Printf("  1. 进入项目目录: cd %s\n", projectDir)
	fmt.Println("  2. 配置数据库连接: ora2pg-admin 配置 数据库")
	fmt.Println("  3. 检查环境: ora2pg-admin 检查 环境")
	fmt.Println("  4. 测试连接: ora2pg-admin 检查 连接")
	fmt.Println("  5. 执行迁移: ora2pg-admin 迁移 全部")

	fmt.Println()
	fmt.Println("💡 提示:")
	fmt.Println("  • 使用 'ora2pg-admin 帮助' 查看所有可用命令")
	fmt.Println("  • 配置文件位于 .ora2pg-admin/config.yaml")
	fmt.Println("  • 查看 README.md 了解更多信息")

	fmt.Println()
	fmt.Println("✨ 祝您迁移顺利！")
}
