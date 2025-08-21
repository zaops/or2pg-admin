package service

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOra2pgService(t *testing.T) {
	service := NewOra2pgService()
	assert.NotNil(t, service)
	assert.NotNil(t, service.logger)
	assert.NotNil(t, service.fileUtils)
}

func TestGetSupportedTypes(t *testing.T) {
	service := NewOra2pgService()
	types := service.GetSupportedTypes()
	
	assert.NotEmpty(t, types)
	assert.Contains(t, types, MigrationTypeTable)
	assert.Contains(t, types, MigrationTypeView)
	assert.Contains(t, types, MigrationTypeSequence)
	assert.Contains(t, types, MigrationTypeIndex)
	assert.Contains(t, types, MigrationTypeTrigger)
	assert.Contains(t, types, MigrationTypeFunction)
	assert.Contains(t, types, MigrationTypeProcedure)
	assert.Contains(t, types, MigrationTypePackage)
	assert.Contains(t, types, MigrationTypeType)
	assert.Contains(t, types, MigrationTypeGrant)
	assert.Contains(t, types, MigrationTypeCopy)
	assert.Contains(t, types, MigrationTypeInsert)
}

func TestValidateMigrationType(t *testing.T) {
	service := NewOra2pgService()
	
	// 测试有效类型
	validTypes := []MigrationType{
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
	
	for _, validType := range validTypes {
		err := service.ValidateMigrationType(validType)
		assert.NoError(t, err, "类型 %s 应该是有效的", validType)
	}
	
	// 测试无效类型
	invalidTypes := []MigrationType{
		"INVALID_TYPE",
		"UNKNOWN",
		"",
	}
	
	for _, invalidType := range invalidTypes {
		err := service.ValidateMigrationType(invalidType)
		assert.Error(t, err, "类型 %s 应该是无效的", invalidType)
	}
}

func TestExecutionOptions(t *testing.T) {
	options := &ExecutionOptions{
		ConfigFile:  "test.conf",
		OutputDir:   "output",
		LogFile:     "test.log",
		DryRun:      true,
		Verbose:     true,
		Timeout:     30 * time.Second,
		WorkingDir:  ".",
		Environment: map[string]string{
			"ORACLE_HOME": "/opt/oracle",
			"NLS_LANG":    "AMERICAN_AMERICA.UTF8",
		},
	}
	
	assert.Equal(t, "test.conf", options.ConfigFile)
	assert.Equal(t, "output", options.OutputDir)
	assert.Equal(t, "test.log", options.LogFile)
	assert.True(t, options.DryRun)
	assert.True(t, options.Verbose)
	assert.Equal(t, 30*time.Second, options.Timeout)
	assert.Equal(t, ".", options.WorkingDir)
	assert.Equal(t, "/opt/oracle", options.Environment["ORACLE_HOME"])
	assert.Equal(t, "AMERICAN_AMERICA.UTF8", options.Environment["NLS_LANG"])
}

func TestExecutionResult(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(5 * time.Minute)
	
	result := &ExecutionResult{
		Status:      StatusCompleted,
		StartTime:   startTime,
		EndTime:     endTime,
		Duration:    endTime.Sub(startTime),
		ExitCode:    0,
		Output:      "Migration completed successfully",
		ErrorOutput: "",
		Progress: &ProgressInfo{
			CurrentStep:    "完成",
			TotalSteps:     10,
			CompletedSteps: 10,
			Percentage:     100.0,
			ProcessedRows:  1000,
			TotalRows:      1000,
			Message:        "迁移完成",
		},
		Error: nil,
	}
	
	assert.Equal(t, StatusCompleted, result.Status)
	assert.Equal(t, startTime, result.StartTime)
	assert.Equal(t, endTime, result.EndTime)
	assert.Equal(t, 5*time.Minute, result.Duration)
	assert.Equal(t, 0, result.ExitCode)
	assert.Equal(t, "Migration completed successfully", result.Output)
	assert.Empty(t, result.ErrorOutput)
	assert.NotNil(t, result.Progress)
	assert.Equal(t, 100.0, result.Progress.Percentage)
	assert.Nil(t, result.Error)
}

func TestProgressInfo(t *testing.T) {
	progress := &ProgressInfo{
		CurrentStep:    "处理表: USERS",
		TotalSteps:     5,
		CompletedSteps: 3,
		Percentage:     60.0,
		ProcessedRows:  600,
		TotalRows:      1000,
		Message:        "正在处理用户表",
	}
	
	assert.Equal(t, "处理表: USERS", progress.CurrentStep)
	assert.Equal(t, 5, progress.TotalSteps)
	assert.Equal(t, 3, progress.CompletedSteps)
	assert.Equal(t, 60.0, progress.Percentage)
	assert.Equal(t, int64(600), progress.ProcessedRows)
	assert.Equal(t, int64(1000), progress.TotalRows)
	assert.Equal(t, "正在处理用户表", progress.Message)
}

func TestGetExecutionSummary(t *testing.T) {
	service := NewOra2pgService()
	
	// 创建测试结果
	results := []*ExecutionResult{
		{
			Status:    StatusCompleted,
			StartTime: time.Now().Add(-10 * time.Minute),
			EndTime:   time.Now().Add(-8 * time.Minute),
			Duration:  2 * time.Minute,
			ExitCode:  0,
		},
		{
			Status:    StatusFailed,
			StartTime: time.Now().Add(-8 * time.Minute),
			EndTime:   time.Now().Add(-7 * time.Minute),
			Duration:  1 * time.Minute,
			ExitCode:  1,
			Error:     assert.AnError,
		},
		{
			Status:    StatusCompleted,
			StartTime: time.Now().Add(-7 * time.Minute),
			EndTime:   time.Now().Add(-5 * time.Minute),
			Duration:  2 * time.Minute,
			ExitCode:  0,
		},
		{
			Status:    StatusCancelled,
			StartTime: time.Now().Add(-5 * time.Minute),
			EndTime:   time.Now().Add(-4 * time.Minute),
			Duration:  1 * time.Minute,
			ExitCode:  -1,
		},
	}
	
	summary := service.GetExecutionSummary(results)
	
	assert.Equal(t, 4, summary["total_executions"])
	assert.Equal(t, 2, summary["successful"])
	assert.Equal(t, 1, summary["failed"])
	assert.Equal(t, 1, summary["cancelled"])
	assert.Equal(t, 6*time.Minute, summary["total_duration"])
	
	details := summary["details"].([]map[string]interface{})
	assert.Len(t, details, 4)
	
	// 检查第一个结果的详情
	firstDetail := details[0]
	assert.Equal(t, StatusCompleted, firstDetail["status"])
	assert.Equal(t, 2*time.Minute, firstDetail["duration"])
	assert.Equal(t, 0, firstDetail["exit_code"])
	
	// 检查失败结果的详情
	failedDetail := details[1]
	assert.Equal(t, StatusFailed, failedDetail["status"])
	assert.Equal(t, 1*time.Minute, failedDetail["duration"])
	assert.Equal(t, 1, failedDetail["exit_code"])
	assert.Contains(t, failedDetail, "error")
}

func TestBuildCommandArgs(t *testing.T) {
	service := NewOra2pgService()

	// 创建临时配置文件
	tempFile, err := os.CreateTemp("", "test-config-*.conf")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	options := &ExecutionOptions{
		ConfigFile: tempFile.Name(),
		OutputDir:  "output",
		LogFile:    "test.log",
		DryRun:     true,
		Verbose:    true,
	}

	args, err := service.buildCommandArgs(MigrationTypeTable, options)
	require.NoError(t, err)

	// 验证基本参数
	assert.Contains(t, args, "ora2pg")
	assert.Contains(t, args, "-c")
	assert.Contains(t, args, tempFile.Name())
	assert.Contains(t, args, "-t")
	assert.Contains(t, args, "TABLE")
	assert.Contains(t, args, "-o")
	assert.Contains(t, args, "output")
	assert.Contains(t, args, "-l")
	assert.Contains(t, args, "test.log")
	assert.Contains(t, args, "-v")
	assert.Contains(t, args, "-n")
}

func TestParseProgress(t *testing.T) {
	service := NewOra2pgService()
	progress := &ProgressInfo{}
	
	testCases := []struct {
		line     string
		expected map[string]interface{}
	}{
		{
			line: "Processing table: USERS (3/10)",
			expected: map[string]interface{}{
				"current_step":     "处理table: USERS",
				"completed_steps":  3,
				"total_steps":      10,
				"percentage":       30.0,
			},
		},
		{
			line: "Exported 1500 rows",
			expected: map[string]interface{}{
				"processed_rows": int64(1500),
				"message":        "已导出 1500 行数据",
			},
		},
		{
			line: "Total rows: 5000",
			expected: map[string]interface{}{
				"total_rows": int64(5000),
			},
		},
		{
			line: "INFO: Starting migration process",
			expected: map[string]interface{}{
				"message": "Starting migration process",
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.line, func(t *testing.T) {
			// 重置进度信息
			progress = &ProgressInfo{}
			
			// 解析进度
			service.parseProgress(tc.line, progress)
			
			// 验证结果
			for key, expectedValue := range tc.expected {
				switch key {
				case "current_step":
					assert.Equal(t, expectedValue, progress.CurrentStep)
				case "completed_steps":
					assert.Equal(t, expectedValue, progress.CompletedSteps)
				case "total_steps":
					assert.Equal(t, expectedValue, progress.TotalSteps)
				case "percentage":
					assert.Equal(t, expectedValue, progress.Percentage)
				case "processed_rows":
					assert.Equal(t, expectedValue, progress.ProcessedRows)
				case "total_rows":
					assert.Equal(t, expectedValue, progress.TotalRows)
				case "message":
					assert.Equal(t, expectedValue, progress.Message)
				}
			}
		})
	}
}

func TestIsImportantLogLine(t *testing.T) {
	service := NewOra2pgService()
	
	importantLines := []string{
		"ERROR: Connection failed",
		"WARNING: Table not found",
		"FATAL: Critical error occurred",
		"Processing table USERS",
		"Exported 1000 rows",
		"Total rows: 5000",
		"Completed successfully",
		"Failed to process",
	}
	
	unimportantLines := []string{
		"Debug: Internal state",
		"Trace: Function entry",
		"Random log message",
		"",
	}
	
	for _, line := range importantLines {
		assert.True(t, service.isImportantLogLine(line), "Line should be important: %s", line)
	}
	
	for _, line := range unimportantLines {
		assert.False(t, service.isImportantLogLine(line), "Line should not be important: %s", line)
	}
}

func TestExecuteWithInvalidTool(t *testing.T) {
	service := NewOra2pgService()
	
	ctx := context.Background()
	options := &ExecutionOptions{
		ConfigFile: "test.conf",
		OutputDir:  "output",
		Timeout:    5 * time.Second,
	}
	
	// 这个测试会失败，因为系统中没有安装ora2pg工具
	result, err := service.Execute(ctx, MigrationTypeTable, options)
	
	// 应该返回错误，因为ora2pg工具不存在
	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, StatusFailed, result.Status)
	assert.NotNil(t, result.Error)
}
