package oracle

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// ClientInfo Oracle客户端信息
type ClientInfo struct {
	Installed    bool   `json:"installed"`
	Version      string `json:"version"`
	Home         string `json:"home"`
	InstantClient bool   `json:"instant_client"`
	Architecture string `json:"architecture"`
	Path         string `json:"path"`
}

// InstallationGuide 安装指导信息
type InstallationGuide struct {
	Platform     string   `json:"platform"`
	DownloadURL  string   `json:"download_url"`
	Instructions []string `json:"instructions"`
}

// ClientDetector Oracle客户端检测器
type ClientDetector struct {
	clientInfo *ClientInfo
}

// NewClientDetector 创建新的客户端检测器
func NewClientDetector() *ClientDetector {
	return &ClientDetector{
		clientInfo: &ClientInfo{},
	}
}

// DetectClient 检测Oracle客户端
func (cd *ClientDetector) DetectClient() (*ClientInfo, error) {
	logrus.Debug("开始检测Oracle客户端...")

	// 重置客户端信息
	cd.clientInfo = &ClientInfo{
		Architecture: runtime.GOARCH,
	}

	// 1. 检查ORACLE_HOME环境变量
	if oracleHome := os.Getenv("ORACLE_HOME"); oracleHome != "" {
		logrus.Debugf("发现ORACLE_HOME环境变量: %s", oracleHome)
		cd.clientInfo.Home = oracleHome
		if cd.validateOracleHome(oracleHome) {
			cd.clientInfo.Installed = true
			cd.clientInfo.InstantClient = false
			cd.detectVersion()
			return cd.clientInfo, nil
		}
	}

	// 2. 检查常见的Oracle客户端安装路径
	commonPaths := cd.getCommonOraclePaths()
	for _, path := range commonPaths {
		if cd.validateOracleHome(path) {
			logrus.Debugf("在路径 %s 发现Oracle客户端", path)
			cd.clientInfo.Home = path
			cd.clientInfo.Installed = true
			cd.clientInfo.InstantClient = strings.Contains(strings.ToLower(path), "instantclient")
			cd.detectVersion()
			return cd.clientInfo, nil
		}
	}

	// 3. 检查PATH中的Oracle工具
	if cd.checkOracleInPath() {
		cd.clientInfo.Installed = true
		cd.detectVersion()
		return cd.clientInfo, nil
	}

	// 4. 未找到Oracle客户端
	logrus.Warn("未检测到Oracle客户端")
	cd.clientInfo.Installed = false
	return cd.clientInfo, nil
}

// validateOracleHome 验证Oracle Home目录
func (cd *ClientDetector) validateOracleHome(oracleHome string) bool {
	if oracleHome == "" {
		return false
	}

	// 检查目录是否存在
	if _, err := os.Stat(oracleHome); os.IsNotExist(err) {
		return false
	}

	// 检查关键文件
	requiredFiles := []string{
		"bin/sqlplus" + cd.getExecutableExtension(),
		"lib",
	}

	// 对于Instant Client，检查不同的文件
	if strings.Contains(strings.ToLower(oracleHome), "instantclient") {
		requiredFiles = []string{
			"sqlplus" + cd.getExecutableExtension(),
		}
	}

	for _, file := range requiredFiles {
		fullPath := filepath.Join(oracleHome, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			logrus.Debugf("未找到必需文件: %s", fullPath)
			return false
		}
	}

	return true
}

// getCommonOraclePaths 获取常见的Oracle安装路径
func (cd *ClientDetector) getCommonOraclePaths() []string {
	var paths []string

	switch runtime.GOOS {
	case "windows":
		// Windows常见路径
		drives := []string{"C:", "D:", "E:"}
		for _, drive := range drives {
			paths = append(paths, []string{
				filepath.Join(drive, "app", "oracle", "product"),
				filepath.Join(drive, "oracle", "product"),
				filepath.Join(drive, "Oracle", "instantclient"),
				filepath.Join(drive, "instantclient"),
				filepath.Join(drive, "Program Files", "Oracle"),
				filepath.Join(drive, "Program Files (x86)", "Oracle"),
			}...)
		}
	case "linux":
		// Linux常见路径
		paths = []string{
			"/opt/oracle",
			"/usr/lib/oracle",
			"/home/oracle",
			"/opt/instantclient",
			"/usr/local/oracle",
		}
	case "darwin":
		// macOS常见路径
		paths = []string{
			"/opt/oracle",
			"/usr/local/oracle",
			"/Applications/Oracle",
			"/opt/instantclient",
		}
	}

	// 扩展路径，查找子目录
	var expandedPaths []string
	for _, basePath := range paths {
		if entries, err := os.ReadDir(basePath); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					subPath := filepath.Join(basePath, entry.Name())
					expandedPaths = append(expandedPaths, subPath)
					
					// 进一步查找版本目录
					if subEntries, err := os.ReadDir(subPath); err == nil {
						for _, subEntry := range subEntries {
							if subEntry.IsDir() {
								expandedPaths = append(expandedPaths, filepath.Join(subPath, subEntry.Name()))
							}
						}
					}
				}
			}
		}
	}

	return append(paths, expandedPaths...)
}

// checkOracleInPath 检查PATH中的Oracle工具
func (cd *ClientDetector) checkOracleInPath() bool {
	tools := []string{"sqlplus", "tnsping", "lsnrctl"}
	
	for _, tool := range tools {
		if path, err := exec.LookPath(tool); err == nil {
			logrus.Debugf("在PATH中发现Oracle工具: %s -> %s", tool, path)
			cd.clientInfo.Path = filepath.Dir(path)
			return true
		}
	}
	
	return false
}

// detectVersion 检测Oracle客户端版本
func (cd *ClientDetector) detectVersion() {
	// 尝试通过sqlplus获取版本信息
	var sqlplusPath string
	
	if cd.clientInfo.Home != "" {
		if cd.clientInfo.InstantClient {
			sqlplusPath = filepath.Join(cd.clientInfo.Home, "sqlplus"+cd.getExecutableExtension())
		} else {
			sqlplusPath = filepath.Join(cd.clientInfo.Home, "bin", "sqlplus"+cd.getExecutableExtension())
		}
	} else if cd.clientInfo.Path != "" {
		sqlplusPath = filepath.Join(cd.clientInfo.Path, "sqlplus"+cd.getExecutableExtension())
	} else {
		// 尝试从PATH中查找
		if path, err := exec.LookPath("sqlplus"); err == nil {
			sqlplusPath = path
		}
	}

	if sqlplusPath == "" {
		logrus.Debug("未找到sqlplus，无法检测版本")
		return
	}

	// 执行sqlplus -version命令
	cmd := exec.Command(sqlplusPath, "-version")
	output, err := cmd.Output()
	if err != nil {
		logrus.Debugf("执行sqlplus -version失败: %v", err)
		return
	}

	// 解析版本信息
	version := cd.parseVersion(string(output))
	if version != "" {
		cd.clientInfo.Version = version
		logrus.Debugf("检测到Oracle客户端版本: %s", version)
	}
}

// parseVersion 解析版本信息
func (cd *ClientDetector) parseVersion(output string) string {
	// 匹配版本号模式，如 "Release 19.0.0.0.0" 或 "Version 12.2.0.1.0"
	patterns := []string{
		`Release\s+(\d+\.\d+\.\d+\.\d+\.\d+)`,
		`Version\s+(\d+\.\d+\.\d+\.\d+\.\d+)`,
		`(\d+\.\d+\.\d+\.\d+\.\d+)`,
		`(\d+\.\d+\.\d+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(output); len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

// getExecutableExtension 获取可执行文件扩展名
func (cd *ClientDetector) getExecutableExtension() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

// GetClientInfo 获取客户端信息
func (cd *ClientDetector) GetClientInfo() *ClientInfo {
	return cd.clientInfo
}

// IsCompatible 检查版本兼容性
func (cd *ClientDetector) IsCompatible(version string) bool {
	if version == "" {
		return false
	}

	// 提取主版本号
	parts := strings.Split(version, ".")
	if len(parts) == 0 {
		return false
	}

	majorVersion := parts[0]
	
	// 支持的Oracle版本：11g, 12c, 18c, 19c, 21c
	supportedVersions := []string{"11", "12", "18", "19", "21"}
	
	for _, supported := range supportedVersions {
		if majorVersion == supported {
			return true
		}
	}

	return false
}

// GetInstallationGuide 获取安装指导
func (cd *ClientDetector) GetInstallationGuide() *InstallationGuide {
	guide := &InstallationGuide{
		Platform: runtime.GOOS,
	}

	switch runtime.GOOS {
	case "windows":
		guide.DownloadURL = "https://www.oracle.com/database/technologies/instant-client/winx64-64-downloads.html"
		guide.Instructions = []string{
			"1. 访问Oracle官方下载页面",
			"2. 下载适合您系统的Instant Client",
			"3. 解压到目录（如 C:\\instantclient）",
			"4. 将解压目录添加到PATH环境变量",
			"5. 设置ORACLE_HOME环境变量指向解压目录",
			"6. 重启命令行工具以使环境变量生效",
		}
	case "linux":
		guide.DownloadURL = "https://www.oracle.com/database/technologies/instant-client/linux-x86-64-downloads.html"
		guide.Instructions = []string{
			"1. 访问Oracle官方下载页面",
			"2. 下载适合您系统的Instant Client RPM或ZIP包",
			"3. 安装或解压到目录（如 /opt/instantclient）",
			"4. 设置环境变量：",
			"   export ORACLE_HOME=/opt/instantclient",
			"   export PATH=$ORACLE_HOME:$PATH",
			"   export LD_LIBRARY_PATH=$ORACLE_HOME:$LD_LIBRARY_PATH",
			"5. 将环境变量添加到 ~/.bashrc 或 ~/.profile",
		}
	case "darwin":
		guide.DownloadURL = "https://www.oracle.com/database/technologies/instant-client/macos-intel-x86-downloads.html"
		guide.Instructions = []string{
			"1. 访问Oracle官方下载页面",
			"2. 下载适合您系统的Instant Client",
			"3. 解压到目录（如 /opt/instantclient）",
			"4. 设置环境变量：",
			"   export ORACLE_HOME=/opt/instantclient",
			"   export PATH=$ORACLE_HOME:$PATH",
			"   export DYLD_LIBRARY_PATH=$ORACLE_HOME:$DYLD_LIBRARY_PATH",
			"5. 将环境变量添加到 ~/.zshrc 或 ~/.bash_profile",
		}
	}

	return guide
}

// CheckClientStatus 检查客户端状态
func (cd *ClientDetector) CheckClientStatus() *ClientStatusReport {
	report := &ClientStatusReport{
		Timestamp: time.Now(),
	}

	// 检测客户端
	clientInfo, err := cd.DetectClient()
	if err != nil {
		report.Status = "ERROR"
		report.Message = fmt.Sprintf("检测失败: %v", err)
		return report
	}

	report.ClientInfo = *clientInfo

	if !clientInfo.Installed {
		report.Status = "NOT_INSTALLED"
		report.Message = "未检测到Oracle客户端"
		report.Recommendations = []string{
			"请安装Oracle Instant Client或完整的Oracle客户端",
			"设置ORACLE_HOME环境变量",
			"将Oracle客户端路径添加到PATH环境变量",
		}
		return report
	}

	// 检查版本兼容性
	if clientInfo.Version != "" {
		if cd.IsCompatible(clientInfo.Version) {
			report.Status = "COMPATIBLE"
			report.Message = fmt.Sprintf("Oracle客户端 %s 已安装且兼容", clientInfo.Version)
		} else {
			report.Status = "INCOMPATIBLE"
			report.Message = fmt.Sprintf("Oracle客户端 %s 版本可能不兼容", clientInfo.Version)
			report.Recommendations = []string{
				"建议使用Oracle 11g、12c、18c、19c或21c版本",
				"请考虑升级到支持的版本",
			}
		}
	} else {
		report.Status = "UNKNOWN_VERSION"
		report.Message = "Oracle客户端已安装，但无法确定版本"
		report.Recommendations = []string{
			"请检查sqlplus工具是否可用",
			"验证Oracle客户端安装是否完整",
		}
	}

	return report
}

// ClientStatusReport 客户端状态报告
type ClientStatusReport struct {
	Timestamp       time.Time   `json:"timestamp"`
	Status          string      `json:"status"` // NOT_INSTALLED, COMPATIBLE, INCOMPATIBLE, UNKNOWN_VERSION, ERROR
	Message         string      `json:"message"`
	ClientInfo      ClientInfo  `json:"client_info"`
	Recommendations []string    `json:"recommendations,omitempty"`
}

// GetStatusSummary 获取状态摘要
func (csr *ClientStatusReport) GetStatusSummary() string {
	var summary strings.Builder

	summary.WriteString("📊 Oracle客户端状态报告\n")
	summary.WriteString(fmt.Sprintf("🕒 检查时间: %s\n", csr.Timestamp.Format("2006-01-02 15:04:05")))
	summary.WriteString("\n")

	switch csr.Status {
	case "COMPATIBLE":
		summary.WriteString("✅ " + csr.Message + "\n")
	case "NOT_INSTALLED":
		summary.WriteString("❌ " + csr.Message + "\n")
	case "INCOMPATIBLE":
		summary.WriteString("⚠️ " + csr.Message + "\n")
	case "UNKNOWN_VERSION":
		summary.WriteString("❓ " + csr.Message + "\n")
	case "ERROR":
		summary.WriteString("💥 " + csr.Message + "\n")
	}

	// 显示客户端详细信息
	if csr.ClientInfo.Installed {
		summary.WriteString("\n📋 客户端详细信息:\n")
		if csr.ClientInfo.Version != "" {
			summary.WriteString(fmt.Sprintf("  版本: %s\n", csr.ClientInfo.Version))
		}
		if csr.ClientInfo.Home != "" {
			summary.WriteString(fmt.Sprintf("  安装路径: %s\n", csr.ClientInfo.Home))
		}
		if csr.ClientInfo.InstantClient {
			summary.WriteString("  类型: Instant Client\n")
		} else {
			summary.WriteString("  类型: 完整客户端\n")
		}
		summary.WriteString(fmt.Sprintf("  架构: %s\n", csr.ClientInfo.Architecture))
	}

	// 显示建议
	if len(csr.Recommendations) > 0 {
		summary.WriteString("\n💡 建议:\n")
		for i, rec := range csr.Recommendations {
			summary.WriteString(fmt.Sprintf("  %d. %s\n", i+1, rec))
		}
	}

	return summary.String()
}
