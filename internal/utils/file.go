package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

// FileUtils 文件操作工具
type FileUtils struct{}

// NewFileUtils 创建文件工具实例
func NewFileUtils() *FileUtils {
	return &FileUtils{}
}

// EnsureDir 确保目录存在，如果不存在则创建
func (fu *FileUtils) EnsureDir(dirPath string) error {
	if dirPath == "" {
		return fmt.Errorf("目录路径不能为空")
	}

	// 检查目录是否已存在
	if info, err := os.Stat(dirPath); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("路径 %s 已存在但不是目录", dirPath)
		}
		logrus.Debugf("目录已存在: %s", dirPath)
		return nil
	}

	// 创建目录（包括父目录）
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("创建目录失败 %s: %v", dirPath, err)
	}

	logrus.Debugf("成功创建目录: %s", dirPath)
	return nil
}

// WriteFile 写入文件内容
func (fu *FileUtils) WriteFile(filePath string, content []byte, perm os.FileMode) error {
	if filePath == "" {
		return fmt.Errorf("文件路径不能为空")
	}

	// 确保父目录存在
	dir := filepath.Dir(filePath)
	if err := fu.EnsureDir(dir); err != nil {
		return fmt.Errorf("创建父目录失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(filePath, content, perm); err != nil {
		return fmt.Errorf("写入文件失败 %s: %v", filePath, err)
	}

	logrus.Debugf("成功写入文件: %s (%d bytes)", filePath, len(content))
	return nil
}

// ReadFile 读取文件内容
func (fu *FileUtils) ReadFile(filePath string) ([]byte, error) {
	if filePath == "" {
		return nil, fmt.Errorf("文件路径不能为空")
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("文件不存在: %s", filePath)
	}

	// 读取文件
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败 %s: %v", filePath, err)
	}

	logrus.Debugf("成功读取文件: %s (%d bytes)", filePath, len(content))
	return content, nil
}

// CopyFile 复制文件
func (fu *FileUtils) CopyFile(srcPath, dstPath string) error {
	if srcPath == "" || dstPath == "" {
		return fmt.Errorf("源文件路径和目标文件路径不能为空")
	}

	// 打开源文件
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("打开源文件失败 %s: %v", srcPath, err)
	}
	defer srcFile.Close()

	// 获取源文件信息
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("获取源文件信息失败: %v", err)
	}

	// 确保目标目录存在
	dstDir := filepath.Dir(dstPath)
	if err := fu.EnsureDir(dstDir); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	// 创建目标文件
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("创建目标文件失败 %s: %v", dstPath, err)
	}
	defer dstFile.Close()

	// 复制文件内容
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("复制文件内容失败: %v", err)
	}

	// 设置文件权限
	if err := os.Chmod(dstPath, srcInfo.Mode()); err != nil {
		logrus.Warnf("设置文件权限失败 %s: %v", dstPath, err)
	}

	logrus.Debugf("成功复制文件: %s -> %s", srcPath, dstPath)
	return nil
}

// FileExists 检查文件是否存在
func (fu *FileUtils) FileExists(filePath string) bool {
	if filePath == "" {
		return false
	}

	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// DirExists 检查目录是否存在
func (fu *FileUtils) DirExists(dirPath string) bool {
	if dirPath == "" {
		return false
	}

	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// RemoveFile 删除文件
func (fu *FileUtils) RemoveFile(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("文件路径不能为空")
	}

	if !fu.FileExists(filePath) {
		logrus.Debugf("文件不存在，无需删除: %s", filePath)
		return nil
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("删除文件失败 %s: %v", filePath, err)
	}

	logrus.Debugf("成功删除文件: %s", filePath)
	return nil
}

// RemoveDir 删除目录（包括内容）
func (fu *FileUtils) RemoveDir(dirPath string) error {
	if dirPath == "" {
		return fmt.Errorf("目录路径不能为空")
	}

	if !fu.DirExists(dirPath) {
		logrus.Debugf("目录不存在，无需删除: %s", dirPath)
		return nil
	}

	if err := os.RemoveAll(dirPath); err != nil {
		return fmt.Errorf("删除目录失败 %s: %v", dirPath, err)
	}

	logrus.Debugf("成功删除目录: %s", dirPath)
	return nil
}

// GetFileSize 获取文件大小
func (fu *FileUtils) GetFileSize(filePath string) (int64, error) {
	if filePath == "" {
		return 0, fmt.Errorf("文件路径不能为空")
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("获取文件信息失败 %s: %v", filePath, err)
	}

	return info.Size(), nil
}

// SetFilePermission 设置文件权限
func (fu *FileUtils) SetFilePermission(filePath string, perm os.FileMode) error {
	if filePath == "" {
		return fmt.Errorf("文件路径不能为空")
	}

	if err := os.Chmod(filePath, perm); err != nil {
		return fmt.Errorf("设置文件权限失败 %s: %v", filePath, err)
	}

	logrus.Debugf("成功设置文件权限: %s (%v)", filePath, perm)
	return nil
}

// GetExecutablePath 获取当前可执行文件的路径
func (fu *FileUtils) GetExecutablePath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("获取可执行文件路径失败: %v", err)
	}
	return filepath.Dir(execPath), nil
}

// GetWorkingDir 获取当前工作目录
func (fu *FileUtils) GetWorkingDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取工作目录失败: %v", err)
	}
	return wd, nil
}

// IsAbsolutePath 检查是否为绝对路径
func (fu *FileUtils) IsAbsolutePath(path string) bool {
	return filepath.IsAbs(path)
}

// JoinPath 连接路径
func (fu *FileUtils) JoinPath(paths ...string) string {
	return filepath.Join(paths...)
}

// CleanPath 清理路径
func (fu *FileUtils) CleanPath(path string) string {
	return filepath.Clean(path)
}

// GetFileExtension 获取文件扩展名
func (fu *FileUtils) GetFileExtension(filePath string) string {
	return strings.ToLower(filepath.Ext(filePath))
}

// GetFileName 获取文件名（不含路径）
func (fu *FileUtils) GetFileName(filePath string) string {
	return filepath.Base(filePath)
}

// GetFileNameWithoutExt 获取文件名（不含扩展名）
func (fu *FileUtils) GetFileNameWithoutExt(filePath string) string {
	fileName := filepath.Base(filePath)
	ext := filepath.Ext(fileName)
	return strings.TrimSuffix(fileName, ext)
}

// GetExecutableExtension 获取当前平台的可执行文件扩展名
func (fu *FileUtils) GetExecutableExtension() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}
