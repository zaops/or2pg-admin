package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"ora2pg-admin/internal/config"
	"ora2pg-admin/internal/utils"
)

// MigrationType 迁移类型
type MigrationType string

const (
	MigrationTypeTable     MigrationType = "TABLE"
	MigrationTypeView      MigrationType = "VIEW"
	MigrationTypeSequence  MigrationType = "SEQUENCE"
	MigrationTypeIndex     MigrationType = "INDEX"
	MigrationTypeTrigger   MigrationType = "TRIGGER"
	MigrationTypeFunction  MigrationType = "FUNCTION"
	MigrationTypeProcedure MigrationType = "PROCEDURE"
	MigrationTypePackage   MigrationType = "PACKAGE"
	MigrationTypeType      MigrationType = "TYPE"
	MigrationTypeGrant     MigrationType = "GRANT"
	MigrationTypeCopy      MigrationType = "COPY"
	MigrationTypeInsert    MigrationType = "INSERT"
)

// ExecutionStatus 执行状态
type ExecutionStatus string

const (
	StatusPending    ExecutionStatus = "PENDING"
	StatusRunning    ExecutionStatus = "RUNNING"
	StatusCompleted  ExecutionStatus = "COMPLETED"
	StatusFailed     ExecutionStatus = "FAILED"
	StatusCancelled  ExecutionStatus = "CANCELLED"
)

// ExecutionResult 执行结果
type ExecutionResult struct {
	Status       ExecutionStatus `json:"status"`
	StartTime    time.Time       `json:"start_time"`
	EndTime      time.Time       `json:"end_time"`
	Duration     time.Duration   `json:"duration"`
	ExitCode     int             `json:"exit_code"`
	Output       string          `json:"output"`
	ErrorOutput  string          `json:"error_output"`
	Progress     *ProgressInfo   `json:"progress,omitempty"`
	Error        error           `json:"error,omitempty"`
}

// ProgressInfo 进度信息
type ProgressInfo struct {
	CurrentStep   string  `json:"current_step"`
	TotalSteps    int     `json:"total_steps"`
	CompletedSteps int    `json:"completed_steps"`
	Percentage    float64 `json:"percentage"`
	ProcessedRows int64   `json:"processed_rows"`
	TotalRows     int64   `json:"total_rows"`
	Message       string  `json:"message"`
}

// ExecutionOptions 执行选项
type ExecutionOptions struct {
	ConfigFile    string            `json:"config_file"`
	OutputDir     string            `json:"output_dir"`
	LogFile       string            `json:"log_file"`
	DryRun        bool              `json:"dry_run"`
	Verbose       bool              `json:"verbose"`
	Timeout       time.Duration     `json:"timeout"`
	Environment   map[string]string `json:"environment"`
	WorkingDir    string            `json:"working_dir"`
}

// Ora2pgService ora2pg包装服务
type Ora2pgService struct {
	logger    *utils.Logger
	fileUtils *utils.FileUtils
}

// NewOra2pgService 创建新的ora2pg服务
func NewOra2pgService() *Ora2pgService {
	return &Ora2pgService{
		logger:    utils.GetGlobalLogger(),
		fileUtils: utils.NewFileUtils(),
	}
}

// Execute 执行ora2pg命令
func (s *Ora2pgService) Execute(ctx context.Context, migrationType MigrationType, options *ExecutionOptions) (*ExecutionResult, error) {
	result := &ExecutionResult{
		Status:    StatusPending,
		StartTime: time.Now(),
		Progress:  &ProgressInfo{},
	}

	s.logger.Infof("开始执行ora2pg迁移，类型: %s", migrationType)

	// 1. 验证ora2pg工具可用性
	if err := s.validateOra2pgTool(); err != nil {
		result.Status = StatusFailed
		result.Error = err
		return result, err
	}

	// 2. 构建命令参数
	args, err := s.buildCommandArgs(migrationType, options)
	if err != nil {
		result.Status = StatusFailed
		result.Error = err
		return result, err
	}

	// 3. 准备执行环境
	if err := s.prepareExecutionEnvironment(options); err != nil {
		result.Status = StatusFailed
		result.Error = err
		return result, err
	}

	// 4. 执行命令
	result.Status = StatusRunning
	if err := s.executeCommand(ctx, args, options, result); err != nil {
		result.Status = StatusFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}

	// 5. 处理执行结果
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	
	if result.ExitCode == 0 {
		result.Status = StatusCompleted
		s.logger.Infof("ora2pg执行成功，耗时: %v", result.Duration)
	} else {
		result.Status = StatusFailed
		s.logger.Errorf("ora2pg执行失败，退出码: %d", result.ExitCode)
	}

	return result, nil
}

// ExecuteMultiple 执行多种类型的迁移
func (s *Ora2pgService) ExecuteMultiple(ctx context.Context, migrationTypes []MigrationType, options *ExecutionOptions) ([]*ExecutionResult, error) {
	results := make([]*ExecutionResult, 0, len(migrationTypes))
	
	s.logger.Infof("开始执行多类型迁移，共 %d 种类型", len(migrationTypes))

	for i, migrationType := range migrationTypes {
		s.logger.Infof("执行迁移 %d/%d: %s", i+1, len(migrationTypes), migrationType)
		
		result, err := s.Execute(ctx, migrationType, options)
		results = append(results, result)
		
		if err != nil {
			s.logger.Errorf("迁移类型 %s 执行失败: %v", migrationType, err)
			// 继续执行其他类型，不中断整个流程
		}
	}

	return results, nil
}

// validateOra2pgTool 验证ora2pg工具可用性
func (s *Ora2pgService) validateOra2pgTool() error {
	_, err := exec.LookPath("ora2pg")
	if err != nil {
		return utils.NewError(utils.ErrorTypeSystem, "ORA2PG_NOT_FOUND").
			Message("未找到ora2pg工具").
			Suggestion("请确认ora2pg已正确安装").
			Suggestion("将ora2pg添加到PATH环境变量").
			Build()
	}
	return nil
}

// buildCommandArgs 构建命令参数
func (s *Ora2pgService) buildCommandArgs(migrationType MigrationType, options *ExecutionOptions) ([]string, error) {
	args := []string{"ora2pg"}

	// 添加配置文件参数
	if options.ConfigFile != "" {
		if !s.fileUtils.FileExists(options.ConfigFile) {
			return nil, utils.FileErrors.NotFound(options.ConfigFile)
		}
		args = append(args, "-c", options.ConfigFile)
	}

	// 添加迁移类型参数
	args = append(args, "-t", string(migrationType))

	// 添加输出目录参数
	if options.OutputDir != "" {
		args = append(args, "-o", options.OutputDir)
	}

	// 添加详细输出参数
	if options.Verbose {
		args = append(args, "-v")
	}

	// 添加预览模式参数
	if options.DryRun {
		args = append(args, "-n")
	}

	// 添加日志文件参数
	if options.LogFile != "" {
		args = append(args, "-l", options.LogFile)
	}

	s.logger.Debugf("构建的命令参数: %v", args)
	return args, nil
}

// prepareExecutionEnvironment 准备执行环境
func (s *Ora2pgService) prepareExecutionEnvironment(options *ExecutionOptions) error {
	// 确保输出目录存在
	if options.OutputDir != "" {
		if err := s.fileUtils.EnsureDir(options.OutputDir); err != nil {
			return utils.FileErrors.CreateFailed(options.OutputDir, err)
		}
	}

	// 确保日志文件目录存在
	if options.LogFile != "" {
		logDir := filepath.Dir(options.LogFile)
		if err := s.fileUtils.EnsureDir(logDir); err != nil {
			return utils.FileErrors.CreateFailed(logDir, err)
		}
	}

	return nil
}

// executeCommand 执行命令
func (s *Ora2pgService) executeCommand(ctx context.Context, args []string, options *ExecutionOptions, result *ExecutionResult) error {
	// 创建命令
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)

	// 设置工作目录
	if options.WorkingDir != "" {
		cmd.Dir = options.WorkingDir
	}

	// 设置环境变量
	cmd.Env = os.Environ()
	for key, value := range options.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// 创建管道用于捕获输出
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("创建stdout管道失败: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("创建stderr管道失败: %v", err)
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动ora2pg命令失败: %v", err)
	}

	// 处理输出
	outputChan := make(chan string, 100)
	errorChan := make(chan string, 100)
	doneChan := make(chan bool, 2)

	// 读取标准输出
	go s.readOutput(stdout, outputChan, doneChan, result)
	// 读取错误输出
	go s.readOutput(stderr, errorChan, doneChan, result)

	// 等待命令完成或超时
	var waitErr error
	if options.Timeout > 0 {
		timer := time.NewTimer(options.Timeout)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			cmd.Process.Kill()
			waitErr = ctx.Err()
		case <-timer.C:
			cmd.Process.Kill()
			waitErr = fmt.Errorf("命令执行超时")
		case waitErr = <-func() chan error {
			errChan := make(chan error, 1)
			go func() {
				errChan <- cmd.Wait()
			}()
			return errChan
		}():
		}
	} else {
		waitErr = cmd.Wait()
	}

	// 等待输出读取完成
	<-doneChan
	<-doneChan
	close(outputChan)
	close(errorChan)

	// 收集输出
	var outputBuilder, errorBuilder strings.Builder
	for output := range outputChan {
		outputBuilder.WriteString(output)
	}
	for errorOutput := range errorChan {
		errorBuilder.WriteString(errorOutput)
	}

	result.Output = outputBuilder.String()
	result.ErrorOutput = errorBuilder.String()

	// 获取退出码
	if waitErr != nil {
		if exitError, ok := waitErr.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = -1
		}
		return waitErr
	}

	result.ExitCode = 0
	return nil
}

// readOutput 读取命令输出
func (s *Ora2pgService) readOutput(reader io.Reader, outputChan chan<- string, doneChan chan<- bool, result *ExecutionResult) {
	defer func() { doneChan <- true }()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		outputChan <- line + "\n"

		// 解析进度信息
		s.parseProgress(line, result.Progress)

		// 记录重要日志
		if s.isImportantLogLine(line) {
			s.logger.Info(line)
		}
	}

	if err := scanner.Err(); err != nil {
		s.logger.Errorf("读取输出时发生错误: %v", err)
	}
}

// parseProgress 解析进度信息
func (s *Ora2pgService) parseProgress(line string, progress *ProgressInfo) {
	if progress == nil {
		return
	}

	// 解析不同类型的进度信息
	patterns := []struct {
		regex   *regexp.Regexp
		handler func(matches []string, progress *ProgressInfo)
	}{
		{
			// 匹配 "Processing table: TABLE_NAME (1/10)"
			regexp.MustCompile(`Processing\s+(\w+):\s+(\w+)\s+\((\d+)/(\d+)\)`),
			func(matches []string, progress *ProgressInfo) {
				progress.CurrentStep = fmt.Sprintf("处理%s: %s", matches[1], matches[2])
				if completed, err := strconv.Atoi(matches[3]); err == nil {
					progress.CompletedSteps = completed
				}
				if total, err := strconv.Atoi(matches[4]); err == nil {
					progress.TotalSteps = total
					if total > 0 {
						progress.Percentage = float64(progress.CompletedSteps) / float64(total) * 100
					}
				}
			},
		},
		{
			// 匹配 "Exported 1000 rows"
			regexp.MustCompile(`Exported\s+(\d+)\s+rows`),
			func(matches []string, progress *ProgressInfo) {
				if rows, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					progress.ProcessedRows = rows
				}
				progress.Message = fmt.Sprintf("已导出 %s 行数据", matches[1])
			},
		},
		{
			// 匹配 "Total rows: 10000"
			regexp.MustCompile(`Total\s+rows:\s+(\d+)`),
			func(matches []string, progress *ProgressInfo) {
				if rows, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					progress.TotalRows = rows
				}
			},
		},
		{
			// 匹配一般的状态信息
			regexp.MustCompile(`^(INFO|WARNING|ERROR):\s+(.+)`),
			func(matches []string, progress *ProgressInfo) {
				progress.Message = matches[2]
			},
		},
	}

	for _, pattern := range patterns {
		if matches := pattern.regex.FindStringSubmatch(line); matches != nil {
			pattern.handler(matches, progress)
			break
		}
	}
}

// isImportantLogLine 判断是否为重要日志行
func (s *Ora2pgService) isImportantLogLine(line string) bool {
	importantPatterns := []string{
		"ERROR:",
		"WARNING:",
		"FATAL:",
		"Processing",
		"Exported",
		"Total",
		"Completed",
		"Failed",
	}

	lowerLine := strings.ToLower(line)
	for _, pattern := range importantPatterns {
		if strings.Contains(lowerLine, strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

// GetSupportedTypes 获取支持的迁移类型
func (s *Ora2pgService) GetSupportedTypes() []MigrationType {
	return []MigrationType{
		MigrationTypeTable,
		MigrationTypeView,
		MigrationTypeSequence,
		MigrationTypeIndex,
		MigrationTypeTrigger,
		MigrationTypeFunction,
		MigrationTypeProcedure,
		MigrationTypePackage,
		MigrationTypeType,
		MigrationTypeGrant,
		MigrationTypeCopy,
		MigrationTypeInsert,
	}
}

// ValidateMigrationType 验证迁移类型
func (s *Ora2pgService) ValidateMigrationType(migrationType MigrationType) error {
	supportedTypes := s.GetSupportedTypes()
	for _, supportedType := range supportedTypes {
		if migrationType == supportedType {
			return nil
		}
	}

	return utils.ValidationErrors.InvalidFormat("migration_type", string(migrationType))
}

// GenerateConfigFile 生成ora2pg配置文件
func (s *Ora2pgService) GenerateConfigFile(cfg *config.ProjectConfig, outputPath string) error {
	templateEngine := config.NewTemplateEngine("templates")

	// 检查模板目录
	if !s.fileUtils.DirExists("templates") {
		// 尝试使用可执行文件目录的模板
		if execPath, err := s.fileUtils.GetExecutablePath(); err == nil {
			templateDir := s.fileUtils.JoinPath(execPath, "templates")
			if s.fileUtils.DirExists(templateDir) {
				templateEngine.SetTemplateDir(templateDir)
			} else {
				return utils.NewError(utils.ErrorTypeFile, "TEMPLATE_NOT_FOUND").
					Message("未找到ora2pg配置模板").
					Suggestion("请确认templates目录存在").
					Build()
			}
		}
	}

	return templateEngine.GenerateOra2pgConfig(cfg, outputPath)
}

// GetExecutionSummary 获取执行摘要
func (s *Ora2pgService) GetExecutionSummary(results []*ExecutionResult) map[string]interface{} {
	summary := map[string]interface{}{
		"total_executions": len(results),
		"successful":       0,
		"failed":          0,
		"cancelled":       0,
		"total_duration":  time.Duration(0),
		"details":         []map[string]interface{}{},
	}

	for _, result := range results {
		switch result.Status {
		case StatusCompleted:
			summary["successful"] = summary["successful"].(int) + 1
		case StatusFailed:
			summary["failed"] = summary["failed"].(int) + 1
		case StatusCancelled:
			summary["cancelled"] = summary["cancelled"].(int) + 1
		}

		summary["total_duration"] = summary["total_duration"].(time.Duration) + result.Duration

		detail := map[string]interface{}{
			"status":    result.Status,
			"duration":  result.Duration,
			"exit_code": result.ExitCode,
		}

		if result.Progress != nil {
			detail["progress"] = result.Progress
		}

		if result.Error != nil {
			detail["error"] = result.Error.Error()
		}

		summary["details"] = append(summary["details"].([]map[string]interface{}), detail)
	}

	return summary
}
