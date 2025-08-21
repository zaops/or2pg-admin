package oracle

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"ora2pg-admin/internal/config"
)

// ConnectionResult 连接测试结果
type ConnectionResult struct {
	Success      bool          `json:"success"`
	Message      string        `json:"message"`
	ResponseTime time.Duration `json:"response_time"`
	Error        string        `json:"error,omitempty"`
	Details      string        `json:"details,omitempty"`
}

// ConnectionTester 数据库连接测试器
type ConnectionTester struct {
	clientDetector *ClientDetector
}

// NewConnectionTester 创建新的连接测试器
func NewConnectionTester() *ConnectionTester {
	return &ConnectionTester{
		clientDetector: NewClientDetector(),
	}
}

// TestOracleConnection 测试Oracle数据库连接
func (ct *ConnectionTester) TestOracleConnection(oracleConfig *config.OracleConfig) *ConnectionResult {
	startTime := time.Now()
	result := &ConnectionResult{}

	logrus.Debugf("开始测试Oracle连接: %s:%d", oracleConfig.Host, oracleConfig.Port)

	// 1. 检查Oracle客户端是否可用
	clientInfo, err := ct.clientDetector.DetectClient()
	if err != nil {
		result.Error = fmt.Sprintf("检测Oracle客户端失败: %v", err)
		result.Message = "❌ Oracle客户端检测失败"
		return result
	}

	if !clientInfo.Installed {
		result.Error = "未检测到Oracle客户端"
		result.Message = "❌ 未安装Oracle客户端"
		result.Details = "请安装Oracle Instant Client或完整的Oracle客户端"
		return result
	}

	// 2. 使用tnsping测试网络连通性
	if tnsResult := ct.testTNSPing(oracleConfig); !tnsResult.Success {
		result.Error = tnsResult.Error
		result.Message = "❌ 网络连通性测试失败"
		result.Details = tnsResult.Details
		result.ResponseTime = time.Since(startTime)
		return result
	}

	// 3. 使用sqlplus测试数据库连接
	if sqlResult := ct.testSQLPlusConnection(oracleConfig); !sqlResult.Success {
		result.Error = sqlResult.Error
		result.Message = "❌ 数据库连接测试失败"
		result.Details = sqlResult.Details
		result.ResponseTime = time.Since(startTime)
		return result
	}

	// 连接成功
	result.Success = true
	result.Message = "✅ Oracle数据库连接成功"
	result.ResponseTime = time.Since(startTime)
	result.Details = fmt.Sprintf("连接到 %s:%d，响应时间: %v", 
		oracleConfig.Host, oracleConfig.Port, result.ResponseTime)

	logrus.Infof("Oracle连接测试成功，响应时间: %v", result.ResponseTime)
	return result
}

// testTNSPing 使用tnsping测试网络连通性
func (ct *ConnectionTester) testTNSPing(oracleConfig *config.OracleConfig) *ConnectionResult {
	result := &ConnectionResult{}

	// 构建连接字符串
	var connectString string
	if oracleConfig.Service != "" {
		connectString = fmt.Sprintf("%s:%d/%s", oracleConfig.Host, oracleConfig.Port, oracleConfig.Service)
	} else {
		connectString = fmt.Sprintf("%s:%d/%s", oracleConfig.Host, oracleConfig.Port, oracleConfig.SID)
	}

	// 查找tnsping工具
	tnspingPath, err := ct.findOracleTool("tnsping")
	if err != nil {
		logrus.Debug("未找到tnsping工具，跳过网络连通性测试")
		result.Success = true // 不强制要求tnsping
		return result
	}

	// 执行tnsping命令
	cmd := exec.Command(tnspingPath, connectString)
	output, err := cmd.Output()
	if err != nil {
		result.Error = fmt.Sprintf("tnsping执行失败: %v", err)
		result.Details = string(output)
		return result
	}

	// 解析tnsping输出
	outputStr := string(output)
	if strings.Contains(outputStr, "OK") || strings.Contains(outputStr, "成功") {
		result.Success = true
		result.Message = "网络连通性测试通过"
	} else {
		result.Error = "网络连通性测试失败"
		result.Details = outputStr
	}

	return result
}

// testSQLPlusConnection 使用sqlplus测试数据库连接
func (ct *ConnectionTester) testSQLPlusConnection(oracleConfig *config.OracleConfig) *ConnectionResult {
	result := &ConnectionResult{}

	// 查找sqlplus工具
	sqlplusPath, err := ct.findOracleTool("sqlplus")
	if err != nil {
		result.Error = fmt.Sprintf("未找到sqlplus工具: %v", err)
		return result
	}

	// 构建连接字符串
	var connectString string
	if oracleConfig.Service != "" {
		connectString = fmt.Sprintf("%s/%s@%s:%d/%s",
			oracleConfig.Username, oracleConfig.Password,
			oracleConfig.Host, oracleConfig.Port, oracleConfig.Service)
	} else {
		connectString = fmt.Sprintf("%s/%s@%s:%d/%s",
			oracleConfig.Username, oracleConfig.Password,
			oracleConfig.Host, oracleConfig.Port, oracleConfig.SID)
	}

	// 创建测试SQL脚本
	testSQL := "SELECT 'CONNECTION_TEST_OK' FROM DUAL; EXIT;"

	// 执行sqlplus命令
	cmd := exec.Command(sqlplusPath, "-S", connectString)
	cmd.Stdin = strings.NewReader(testSQL)
	
	output, err := cmd.Output()
	if err != nil {
		result.Error = fmt.Sprintf("sqlplus执行失败: %v", err)
		result.Details = string(output)
		return result
	}

	// 解析sqlplus输出
	outputStr := string(output)
	if strings.Contains(outputStr, "CONNECTION_TEST_OK") {
		result.Success = true
		result.Message = "数据库连接测试通过"
	} else if strings.Contains(outputStr, "ORA-") {
		// 提取Oracle错误信息
		result.Error = ct.extractOracleError(outputStr)
		result.Details = outputStr
	} else {
		result.Error = "数据库连接测试失败"
		result.Details = outputStr
	}

	return result
}

// TestPostgreSQLConnection 测试PostgreSQL数据库连接
func (ct *ConnectionTester) TestPostgreSQLConnection(pgConfig *config.PostgreConfig) *ConnectionResult {
	startTime := time.Now()
	result := &ConnectionResult{}

	logrus.Debugf("开始测试PostgreSQL连接: %s:%d", pgConfig.Host, pgConfig.Port)

	// 查找psql工具
	psqlPath, err := exec.LookPath("psql")
	if err != nil {
		result.Error = "未找到psql工具，请安装PostgreSQL客户端"
		result.Message = "❌ PostgreSQL客户端未安装"
		return result
	}

	// 构建连接字符串
	connectString := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		pgConfig.Username, pgConfig.Password,
		pgConfig.Host, pgConfig.Port, pgConfig.Database)

	// 创建测试SQL
	testSQL := "SELECT 'CONNECTION_TEST_OK';"

	// 执行psql命令
	cmd := exec.Command(psqlPath, connectString, "-c", testSQL, "-t", "-A")
	output, err := cmd.Output()
	if err != nil {
		result.Error = fmt.Sprintf("psql执行失败: %v", err)
		result.Details = string(output)
		result.Message = "❌ PostgreSQL连接失败"
		result.ResponseTime = time.Since(startTime)
		return result
	}

	// 检查输出
	outputStr := strings.TrimSpace(string(output))
	if strings.Contains(outputStr, "CONNECTION_TEST_OK") {
		result.Success = true
		result.Message = "✅ PostgreSQL数据库连接成功"
		result.ResponseTime = time.Since(startTime)
		result.Details = fmt.Sprintf("连接到 %s:%d，响应时间: %v",
			pgConfig.Host, pgConfig.Port, result.ResponseTime)
		logrus.Infof("PostgreSQL连接测试成功，响应时间: %v", result.ResponseTime)
	} else {
		result.Error = "PostgreSQL连接测试失败"
		result.Details = outputStr
		result.Message = "❌ PostgreSQL连接失败"
		result.ResponseTime = time.Since(startTime)
	}

	return result
}

// findOracleTool 查找Oracle工具
func (ct *ConnectionTester) findOracleTool(toolName string) (string, error) {
	// 添加可执行文件扩展名
	if runtime.GOOS == "windows" {
		toolName += ".exe"
	}

	// 1. 首先在PATH中查找
	if path, err := exec.LookPath(toolName); err == nil {
		return path, nil
	}

	// 2. 检测Oracle客户端并在其目录中查找
	clientInfo, err := ct.clientDetector.DetectClient()
	if err != nil || !clientInfo.Installed {
		return "", fmt.Errorf("未检测到Oracle客户端")
	}

	// 3. 在Oracle Home中查找
	if clientInfo.Home != "" {
		var toolPath string
		if clientInfo.InstantClient {
			toolPath = filepath.Join(clientInfo.Home, toolName)
		} else {
			toolPath = filepath.Join(clientInfo.Home, "bin", toolName)
		}

		if _, err := exec.LookPath(toolPath); err == nil {
			return toolPath, nil
		}
	}

	// 4. 在检测到的路径中查找
	if clientInfo.Path != "" {
		toolPath := filepath.Join(clientInfo.Path, toolName)
		if _, err := exec.LookPath(toolPath); err == nil {
			return toolPath, nil
		}
	}

	return "", fmt.Errorf("未找到Oracle工具: %s", toolName)
}

// extractOracleError 提取Oracle错误信息
func (ct *ConnectionTester) extractOracleError(output string) string {
	// 匹配Oracle错误模式
	patterns := []string{
		`(ORA-\d+): (.+)`,
		`(TNS-\d+): (.+)`,
		`(SP2-\d+): (.+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(output); len(matches) >= 3 {
			return fmt.Sprintf("%s: %s", matches[1], strings.TrimSpace(matches[2]))
		}
	}

	// 如果没有匹配到特定错误，返回通用错误信息
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "ORA-") || strings.HasPrefix(line, "TNS-") || strings.HasPrefix(line, "SP2-") {
			return line
		}
	}

	return "数据库连接失败"
}

// GetConnectionDiagnostics 获取连接诊断信息
func (ct *ConnectionTester) GetConnectionDiagnostics(oracleConfig *config.OracleConfig) []string {
	var diagnostics []string

	// 检查Oracle客户端
	clientInfo, err := ct.clientDetector.DetectClient()
	if err != nil || !clientInfo.Installed {
		diagnostics = append(diagnostics, "❌ 未检测到Oracle客户端")
		guide := ct.clientDetector.GetInstallationGuide()
		diagnostics = append(diagnostics, fmt.Sprintf("💡 请访问: %s", guide.DownloadURL))
		return diagnostics
	}

	diagnostics = append(diagnostics, "✅ Oracle客户端已安装")
	if clientInfo.Version != "" {
		diagnostics = append(diagnostics, fmt.Sprintf("📋 版本: %s", clientInfo.Version))
	}
	if clientInfo.Home != "" {
		diagnostics = append(diagnostics, fmt.Sprintf("📁 路径: %s", clientInfo.Home))
	}

	// 检查网络连通性
	diagnostics = append(diagnostics, "")
	diagnostics = append(diagnostics, "🔍 连接诊断建议:")
	diagnostics = append(diagnostics, "1. 检查数据库服务器是否运行")
	diagnostics = append(diagnostics, "2. 验证主机名和端口是否正确")
	diagnostics = append(diagnostics, "3. 确认防火墙设置允许连接")
	diagnostics = append(diagnostics, "4. 检查用户名和密码是否正确")
	diagnostics = append(diagnostics, "5. 验证SID或Service Name是否正确")

	return diagnostics
}
