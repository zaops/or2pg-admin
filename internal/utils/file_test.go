package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileUtils(t *testing.T) {
	fileUtils := NewFileUtils()
	assert.NotNil(t, fileUtils)
}

func TestEnsureDir(t *testing.T) {
	fileUtils := NewFileUtils()
	
	// 创建临时目录用于测试
	tempDir, err := os.MkdirTemp("", "fileutils-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	// 测试创建新目录
	testDir := filepath.Join(tempDir, "test", "nested", "dir")
	err = fileUtils.EnsureDir(testDir)
	assert.NoError(t, err)
	
	// 验证目录是否创建成功
	info, err := os.Stat(testDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
	
	// 测试目录已存在的情况
	err = fileUtils.EnsureDir(testDir)
	assert.NoError(t, err)
	
	// 测试空路径
	err = fileUtils.EnsureDir("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "目录路径不能为空")
}

func TestWriteAndReadFile(t *testing.T) {
	fileUtils := NewFileUtils()
	
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "fileutils-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	// 测试写入文件
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := []byte("这是测试内容\n包含中文字符")
	
	err = fileUtils.WriteFile(testFile, testContent, 0644)
	assert.NoError(t, err)
	
	// 验证文件是否存在
	assert.True(t, fileUtils.FileExists(testFile))
	
	// 测试读取文件
	readContent, err := fileUtils.ReadFile(testFile)
	assert.NoError(t, err)
	assert.Equal(t, testContent, readContent)
	
	// 测试读取不存在的文件
	_, err = fileUtils.ReadFile(filepath.Join(tempDir, "nonexistent.txt"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "文件不存在")
}

func TestCopyFile(t *testing.T) {
	fileUtils := NewFileUtils()
	
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "fileutils-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	// 创建源文件
	srcFile := filepath.Join(tempDir, "source.txt")
	testContent := []byte("测试文件复制内容")
	err = fileUtils.WriteFile(srcFile, testContent, 0644)
	require.NoError(t, err)
	
	// 复制文件
	dstFile := filepath.Join(tempDir, "destination.txt")
	err = fileUtils.CopyFile(srcFile, dstFile)
	assert.NoError(t, err)
	
	// 验证目标文件存在且内容正确
	assert.True(t, fileUtils.FileExists(dstFile))
	
	dstContent, err := fileUtils.ReadFile(dstFile)
	assert.NoError(t, err)
	assert.Equal(t, testContent, dstContent)
	
	// 测试复制不存在的文件
	err = fileUtils.CopyFile("nonexistent.txt", dstFile)
	assert.Error(t, err)
}

func TestFileExists(t *testing.T) {
	fileUtils := NewFileUtils()
	
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "fileutils-test-*")
	require.NoError(t, err)
	tempFile.Close()
	defer os.Remove(tempFile.Name())
	
	// 测试存在的文件
	assert.True(t, fileUtils.FileExists(tempFile.Name()))
	
	// 测试不存在的文件
	assert.False(t, fileUtils.FileExists("nonexistent-file.txt"))
	
	// 测试空路径
	assert.False(t, fileUtils.FileExists(""))
}

func TestDirExists(t *testing.T) {
	fileUtils := NewFileUtils()
	
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "fileutils-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	// 测试存在的目录
	assert.True(t, fileUtils.DirExists(tempDir))
	
	// 测试不存在的目录
	assert.False(t, fileUtils.DirExists(filepath.Join(tempDir, "nonexistent")))
	
	// 测试空路径
	assert.False(t, fileUtils.DirExists(""))
}

func TestRemoveFile(t *testing.T) {
	fileUtils := NewFileUtils()
	
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "fileutils-test-*")
	require.NoError(t, err)
	tempFile.Close()
	
	// 验证文件存在
	assert.True(t, fileUtils.FileExists(tempFile.Name()))
	
	// 删除文件
	err = fileUtils.RemoveFile(tempFile.Name())
	assert.NoError(t, err)
	
	// 验证文件已删除
	assert.False(t, fileUtils.FileExists(tempFile.Name()))
	
	// 测试删除不存在的文件（应该不报错）
	err = fileUtils.RemoveFile("nonexistent.txt")
	assert.NoError(t, err)
}

func TestRemoveDir(t *testing.T) {
	fileUtils := NewFileUtils()
	
	// 创建临时目录和文件
	tempDir, err := os.MkdirTemp("", "fileutils-test-*")
	require.NoError(t, err)
	
	// 在目录中创建文件
	testFile := filepath.Join(tempDir, "test.txt")
	err = fileUtils.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)
	
	// 验证目录存在
	assert.True(t, fileUtils.DirExists(tempDir))
	
	// 删除目录
	err = fileUtils.RemoveDir(tempDir)
	assert.NoError(t, err)
	
	// 验证目录已删除
	assert.False(t, fileUtils.DirExists(tempDir))
}

func TestGetFileSize(t *testing.T) {
	fileUtils := NewFileUtils()
	
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "fileutils-test-*")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())
	
	// 写入测试内容
	testContent := []byte("测试文件大小")
	_, err = tempFile.Write(testContent)
	require.NoError(t, err)
	tempFile.Close()
	
	// 获取文件大小
	size, err := fileUtils.GetFileSize(tempFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, int64(len(testContent)), size)
	
	// 测试不存在的文件
	_, err = fileUtils.GetFileSize("nonexistent.txt")
	assert.Error(t, err)
}

func TestPathOperations(t *testing.T) {
	fileUtils := NewFileUtils()
	
	// 测试路径连接
	path := fileUtils.JoinPath("dir1", "dir2", "file.txt")
	expected := filepath.Join("dir1", "dir2", "file.txt")
	assert.Equal(t, expected, path)
	
	// 测试路径清理
	dirtyPath := "dir1//dir2/../dir3/./file.txt"
	cleanPath := fileUtils.CleanPath(dirtyPath)
	expectedClean := filepath.Clean(dirtyPath)
	assert.Equal(t, expectedClean, cleanPath)
	
	// 测试绝对路径检查
	if runtime.GOOS == "windows" {
		assert.True(t, fileUtils.IsAbsolutePath("C:\\test\\path"))
		assert.False(t, fileUtils.IsAbsolutePath("relative\\path"))
	} else {
		assert.True(t, fileUtils.IsAbsolutePath("/test/path"))
		assert.False(t, fileUtils.IsAbsolutePath("relative/path"))
	}
}

func TestFileNameOperations(t *testing.T) {
	fileUtils := NewFileUtils()
	
	testPath := filepath.Join("dir", "subdir", "test.txt")
	
	// 测试获取文件名
	fileName := fileUtils.GetFileName(testPath)
	assert.Equal(t, "test.txt", fileName)
	
	// 测试获取文件扩展名
	ext := fileUtils.GetFileExtension(testPath)
	assert.Equal(t, ".txt", ext)
	
	// 测试获取不含扩展名的文件名
	nameWithoutExt := fileUtils.GetFileNameWithoutExt(testPath)
	assert.Equal(t, "test", nameWithoutExt)
}

func TestGetExecutableExtension(t *testing.T) {
	fileUtils := NewFileUtils()
	
	ext := fileUtils.GetExecutableExtension()
	if runtime.GOOS == "windows" {
		assert.Equal(t, ".exe", ext)
	} else {
		assert.Equal(t, "", ext)
	}
}

func TestGetWorkingDir(t *testing.T) {
	fileUtils := NewFileUtils()
	
	wd, err := fileUtils.GetWorkingDir()
	assert.NoError(t, err)
	assert.NotEmpty(t, wd)
	assert.True(t, filepath.IsAbs(wd))
}

func TestSetFilePermission(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("跳过Windows上的权限测试")
	}
	
	fileUtils := NewFileUtils()
	
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "fileutils-test-*")
	require.NoError(t, err)
	tempFile.Close()
	defer os.Remove(tempFile.Name())
	
	// 设置文件权限
	err = fileUtils.SetFilePermission(tempFile.Name(), 0755)
	assert.NoError(t, err)
	
	// 验证权限设置
	info, err := os.Stat(tempFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0755), info.Mode().Perm())
}
