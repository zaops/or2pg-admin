package service

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"ora2pg-admin/internal/config"
	"ora2pg-admin/internal/utils"
)

// MigrationPhase 迁移阶段
type MigrationPhase string

const (
	PhaseStructure MigrationPhase = "STRUCTURE"
	PhaseData      MigrationPhase = "DATA"
	PhaseIndex     MigrationPhase = "INDEX"
	PhaseFunction  MigrationPhase = "FUNCTION"
	PhaseGrant     MigrationPhase = "GRANT"
)

// MigrationState 迁移状态
type MigrationState struct {
	CurrentPhase    MigrationPhase    `json:"current_phase"`
	CurrentType     MigrationType     `json:"current_type"`
	TotalSteps      int               `json:"total_steps"`
	CompletedSteps  int               `json:"completed_steps"`
	StartTime       time.Time         `json:"start_time"`
	LastUpdateTime  time.Time         `json:"last_update_time"`
	Results         []*ExecutionResult `json:"results"`
	IsCompleted     bool              `json:"is_completed"`
	IsCancelled     bool              `json:"is_cancelled"`
}

// MigrationService 迁移管理服务
type MigrationService struct {
	config       *config.ProjectConfig
	ora2pgService *Ora2pgService
	logger       *utils.Logger
	fileUtils    *utils.FileUtils
	state        *MigrationState
	parallelJobs int
}

// NewMigrationService 创建新的迁移服务
func NewMigrationService(cfg *config.ProjectConfig) *MigrationService {
	return &MigrationService{
		config:        cfg,
		ora2pgService: NewOra2pgService(),
		logger:        utils.GetGlobalLogger(),
		fileUtils:     utils.NewFileUtils(),
		state: &MigrationState{
			Results: make([]*ExecutionResult, 0),
		},
		parallelJobs: cfg.Migration.ParallelJobs,
	}
}

// ExecuteWithProgress 执行迁移并跟踪进度
func (ms *MigrationService) ExecuteWithProgress(ctx context.Context, migrationTypes []MigrationType, 
	progressTracker *ProgressTracker) ([]*ExecutionResult, error) {
	
	ms.logger.Infof("开始执行迁移，类型数量: %d", len(migrationTypes))
	
	// 初始化状态
	ms.state.TotalSteps = len(migrationTypes)
	ms.state.CompletedSteps = 0
	ms.state.StartTime = time.Now()
	ms.state.LastUpdateTime = time.Now()
	ms.state.IsCompleted = false
	ms.state.IsCancelled = false

	// 准备执行环境
	if err := ms.prepareEnvironment(); err != nil {
		return nil, err
	}

	// 生成ora2pg配置文件
	if err := ms.generateOra2pgConfig(); err != nil {
		ms.logger.Warnf("生成ora2pg配置文件失败: %v", err)
	}

	results := make([]*ExecutionResult, 0, len(migrationTypes))

	// 按阶段执行迁移
	for i, migrationType := range migrationTypes {
		select {
		case <-ctx.Done():
			ms.state.IsCancelled = true
			ms.logger.Info("迁移被用户取消")
			return results, ctx.Err()
		default:
		}

		ms.state.CurrentType = migrationType
		ms.state.CurrentPhase = ms.getPhaseForType(migrationType)
		
		// 更新进度
		progressTracker.UpdateStep(i+1, fmt.Sprintf("执行 %s 迁移", migrationType))

		// 执行单个迁移类型
		result, err := ms.executeSingleMigration(ctx, migrationType)
		results = append(results, result)
		ms.state.Results = append(ms.state.Results, result)

		ms.state.CompletedSteps++
		ms.state.LastUpdateTime = time.Now()

		if err != nil {
			ms.logger.Errorf("迁移类型 %s 执行失败: %v", migrationType, err)
			// 继续执行其他类型，不中断整个流程
		} else {
			ms.logger.Infof("迁移类型 %s 执行成功", migrationType)
		}

		// 更新进度详情
		if result.Progress != nil {
			progressTracker.UpdateProgress(result.Progress.Percentage, result.Progress.Message)
		}
	}

	ms.state.IsCompleted = true
	ms.logger.Info("迁移执行完成")
	
	return results, nil
}

// executeSingleMigration 执行单个迁移类型
func (ms *MigrationService) executeSingleMigration(ctx context.Context, migrationType MigrationType) (*ExecutionResult, error) {
	// 准备执行选项
	options := &ExecutionOptions{
		ConfigFile:  ms.getConfigFilePath(),
		OutputDir:   ms.config.Migration.OutputDir,
		LogFile:     ms.getLogFilePath(migrationType),
		DryRun:      false, // 可以从配置或参数获取
		Verbose:     false,
		Timeout:     30 * time.Minute, // 默认超时时间
		WorkingDir:  ".",
		Environment: ms.buildEnvironment(),
	}

	// 执行ora2pg命令
	return ms.ora2pgService.Execute(ctx, migrationType, options)
}

// prepareEnvironment 准备执行环境
func (ms *MigrationService) prepareEnvironment() error {
	// 确保输出目录存在
	if err := ms.fileUtils.EnsureDir(ms.config.Migration.OutputDir); err != nil {
		return utils.FileErrors.CreateFailed(ms.config.Migration.OutputDir, err)
	}

	// 确保日志目录存在
	logDir := "logs"
	if err := ms.fileUtils.EnsureDir(logDir); err != nil {
		return utils.FileErrors.CreateFailed(logDir, err)
	}

	// 确保备份目录存在
	backupDir := "backup"
	if err := ms.fileUtils.EnsureDir(backupDir); err != nil {
		return utils.FileErrors.CreateFailed(backupDir, err)
	}

	return nil
}

// generateOra2pgConfig 生成ora2pg配置文件
func (ms *MigrationService) generateOra2pgConfig() error {
	configPath := filepath.Join(ms.config.Migration.OutputDir, "ora2pg.conf")
	return ms.ora2pgService.GenerateConfigFile(ms.config, configPath)
}

// getConfigFilePath 获取配置文件路径
func (ms *MigrationService) getConfigFilePath() string {
	return filepath.Join(ms.config.Migration.OutputDir, "ora2pg.conf")
}

// getLogFilePath 获取日志文件路径
func (ms *MigrationService) getLogFilePath(migrationType MigrationType) string {
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("ora2pg-%s-%s.log", migrationType, timestamp)
	return filepath.Join("logs", filename)
}

// buildEnvironment 构建环境变量
func (ms *MigrationService) buildEnvironment() map[string]string {
	env := make(map[string]string)
	
	// 设置Oracle相关环境变量
	if ms.config.OracleClient.Home != "" {
		env["ORACLE_HOME"] = ms.config.OracleClient.Home
	}
	
	// 设置其他必要的环境变量
	env["NLS_LANG"] = "AMERICAN_AMERICA.UTF8"
	
	return env
}

// getPhaseForType 根据迁移类型获取阶段
func (ms *MigrationService) getPhaseForType(migrationType MigrationType) MigrationPhase {
	switch migrationType {
	case MigrationTypeTable, MigrationTypeView, MigrationTypeSequence:
		return PhaseStructure
	case MigrationTypeCopy, MigrationTypeInsert:
		return PhaseData
	case MigrationTypeIndex:
		return PhaseIndex
	case MigrationTypeFunction, MigrationTypeProcedure, MigrationTypeTrigger:
		return PhaseFunction
	case MigrationTypeGrant:
		return PhaseGrant
	default:
		return PhaseStructure
	}
}

// GetState 获取当前迁移状态
func (ms *MigrationService) GetState() *MigrationState {
	return ms.state
}

// SetParallelJobs 设置并行作业数
func (ms *MigrationService) SetParallelJobs(jobs int) {
	ms.parallelJobs = jobs
	if ms.config != nil {
		ms.config.Migration.ParallelJobs = jobs
	}
}

// GetProgress 获取迁移进度
func (ms *MigrationService) GetProgress() float64 {
	if ms.state.TotalSteps == 0 {
		return 0
	}
	return float64(ms.state.CompletedSteps) / float64(ms.state.TotalSteps) * 100
}

// IsCompleted 检查是否完成
func (ms *MigrationService) IsCompleted() bool {
	return ms.state.IsCompleted
}

// IsCancelled 检查是否被取消
func (ms *MigrationService) IsCancelled() bool {
	return ms.state.IsCancelled
}

// GetDuration 获取执行时长
func (ms *MigrationService) GetDuration() time.Duration {
	if ms.state.StartTime.IsZero() {
		return 0
	}
	
	endTime := ms.state.LastUpdateTime
	if ms.state.IsCompleted || ms.state.IsCancelled {
		endTime = ms.state.LastUpdateTime
	} else {
		endTime = time.Now()
	}
	
	return endTime.Sub(ms.state.StartTime)
}
