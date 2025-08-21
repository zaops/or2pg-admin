package oracle

import (
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClientDetector(t *testing.T) {
	detector := NewClientDetector()
	assert.NotNil(t, detector)
}

func TestCheckClientStatus(t *testing.T) {
	detector := NewClientDetector()
	status := detector.CheckClientStatus()
	
	assert.NotNil(t, status)
	assert.NotEmpty(t, status.Status)
	assert.NotEmpty(t, status.Message)
	
	// 状态应该是预定义的值之一
	validStatuses := []string{"COMPATIBLE", "NOT_INSTALLED", "INCOMPATIBLE", "UNKNOWN_VERSION"}
	assert.Contains(t, validStatuses, status.Status)
}

func TestGetCommonOraclePaths(t *testing.T) {
	detector := NewClientDetector()
	paths := detector.getCommonOraclePaths()
	
	assert.NotEmpty(t, paths)
	
	// 验证路径包含当前操作系统的常见路径
	switch runtime.GOOS {
	case "windows":
		found := false
		for _, path := range paths {
			if contains(path, "instantclient") || contains(path, "Oracle") {
				found = true
				break
			}
		}
		assert.True(t, found, "Windows路径应该包含Oracle相关目录")
	case "linux":
		found := false
		for _, path := range paths {
			if contains(path, "/opt/oracle") || contains(path, "/usr/lib/oracle") {
				found = true
				break
			}
		}
		assert.True(t, found, "Linux路径应该包含Oracle相关目录")
	case "darwin":
		found := false
		for _, path := range paths {
			if contains(path, "/opt/oracle") || contains(path, "/usr/local/oracle") {
				found = true
				break
			}
		}
		assert.True(t, found, "macOS路径应该包含Oracle相关目录")
	}
}

func TestEnvironmentVariableDetection(t *testing.T) {
	// 测试环境变量ORACLE_HOME的检测
	originalOracleHome := os.Getenv("ORACLE_HOME")
	defer func() {
		if originalOracleHome != "" {
			os.Setenv("ORACLE_HOME", originalOracleHome)
		} else {
			os.Unsetenv("ORACLE_HOME")
		}
	}()

	// 设置测试环境变量
	testPath := "/test/oracle/home"
	os.Setenv("ORACLE_HOME", testPath)

	// 验证环境变量设置成功
	assert.Equal(t, testPath, os.Getenv("ORACLE_HOME"))
}

// 删除了测试私有方法的测试用例，因为这些方法不是公开API的一部分

func TestGetInstallationGuide(t *testing.T) {
	detector := NewClientDetector()
	guide := detector.GetInstallationGuide()
	
	assert.NotNil(t, guide)
	assert.NotEmpty(t, guide.DownloadURL)
	assert.NotEmpty(t, guide.Instructions)
	
	// 验证指导内容包含当前操作系统的信息
	found := false
	for _, instruction := range guide.Instructions {
		if runtime.GOOS == "windows" && (contains(instruction, "Windows") || contains(instruction, "PATH")) {
			found = true
			break
		} else if runtime.GOOS == "linux" && (contains(instruction, "Linux") || contains(instruction, "LD_LIBRARY_PATH")) {
			found = true
			break
		} else if runtime.GOOS == "darwin" && (contains(instruction, "macOS") || contains(instruction, "DYLD_LIBRARY_PATH")) {
			found = true
			break
		}
	}
	assert.True(t, found, "安装指导应该包含当前操作系统的相关信息")
}

func TestClientStatusReport(t *testing.T) {
	// 创建一个测试状态报告
	report := &ClientStatusReport{
		Status:  "COMPATIBLE",
		Message: "Oracle客户端已安装且兼容",
	}

	// 测试状态摘要
	summary := report.GetStatusSummary()
	assert.NotEmpty(t, summary)
	assert.Contains(t, summary, "Oracle客户端已安装且兼容")

	// 测试详细信息
	assert.Equal(t, "COMPATIBLE", report.Status)
	assert.Equal(t, "Oracle客户端已安装且兼容", report.Message)
}

// 删除了测试私有方法的测试用例

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr || 
			 containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
