package service

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"ora2pg-admin/internal/utils"
)

// ProgressTracker 进度跟踪器
type ProgressTracker struct {
	taskName       string
	totalSteps     int
	currentStep    int
	currentMessage string
	percentage     float64
	startTime      time.Time
	lastUpdateTime time.Time
	isRunning      bool
	mutex          sync.RWMutex
	logger         *utils.Logger
	stopChan       chan bool
	updateChan     chan ProgressUpdate
}

// ProgressUpdate 进度更新信息
type ProgressUpdate struct {
	Step       int
	Message    string
	Percentage float64
	Details    string
}

// NewProgressTracker 创建新的进度跟踪器
func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		logger:     utils.GetGlobalLogger(),
		stopChan:   make(chan bool, 1),
		updateChan: make(chan ProgressUpdate, 100),
	}
}

// Start 开始进度跟踪
func (pt *ProgressTracker) Start(taskName string, totalSteps int) {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()

	pt.taskName = taskName
	pt.totalSteps = totalSteps
	pt.currentStep = 0
	pt.currentMessage = "准备开始..."
	pt.percentage = 0
	pt.startTime = time.Now()
	pt.lastUpdateTime = time.Now()
	pt.isRunning = true

	// 启动进度显示协程
	go pt.displayProgress()

	pt.logger.Infof("开始进度跟踪: %s (总步骤: %d)", taskName, totalSteps)
	fmt.Printf("🚀 开始%s\n", taskName)
	pt.printProgressBar()
}

// Stop 停止进度跟踪
func (pt *ProgressTracker) Stop() {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()

	if !pt.isRunning {
		return
	}

	pt.isRunning = false
	pt.stopChan <- true
	close(pt.updateChan)

	duration := time.Since(pt.startTime)
	pt.logger.Infof("进度跟踪结束: %s (耗时: %v)", pt.taskName, duration)
	
	fmt.Printf("\n✅ %s完成，总耗时: %v\n", pt.taskName, duration)
}

// UpdateStep 更新当前步骤
func (pt *ProgressTracker) UpdateStep(step int, message string) {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()

	if !pt.isRunning {
		return
	}

	pt.currentStep = step
	pt.currentMessage = message
	pt.lastUpdateTime = time.Now()

	if pt.totalSteps > 0 {
		pt.percentage = float64(step) / float64(pt.totalSteps) * 100
	}

	// 发送更新信息
	select {
	case pt.updateChan <- ProgressUpdate{
		Step:       step,
		Message:    message,
		Percentage: pt.percentage,
	}:
	default:
		// 如果通道满了，跳过这次更新
	}

	pt.logger.Debugf("进度更新: 步骤 %d/%d - %s", step, pt.totalSteps, message)
}

// UpdateProgress 更新进度百分比
func (pt *ProgressTracker) UpdateProgress(percentage float64, details string) {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()

	if !pt.isRunning {
		return
	}

	pt.percentage = percentage
	pt.lastUpdateTime = time.Now()

	// 发送更新信息
	select {
	case pt.updateChan <- ProgressUpdate{
		Step:       pt.currentStep,
		Message:    pt.currentMessage,
		Percentage: percentage,
		Details:    details,
	}:
	default:
		// 如果通道满了，跳过这次更新
	}
}

// displayProgress 显示进度（在单独的协程中运行）
func (pt *ProgressTracker) displayProgress() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-pt.stopChan:
			return
		case update := <-pt.updateChan:
			pt.handleProgressUpdate(update)
		case <-ticker.C:
			// 定期刷新显示
			pt.refreshDisplay()
		}
	}
}

// handleProgressUpdate 处理进度更新
func (pt *ProgressTracker) handleProgressUpdate(update ProgressUpdate) {
	fmt.Printf("\r🔄 [%d/%d] %s", update.Step, pt.totalSteps, update.Message)
	
	if update.Details != "" {
		fmt.Printf(" - %s", update.Details)
	}
	
	pt.printProgressBar()
}

// refreshDisplay 刷新显示
func (pt *ProgressTracker) refreshDisplay() {
	pt.mutex.RLock()
	defer pt.mutex.RUnlock()

	if !pt.isRunning {
		return
	}

	elapsed := time.Since(pt.startTime)
	fmt.Printf("\r🔄 [%d/%d] %s (已用时: %v)", 
		pt.currentStep, pt.totalSteps, pt.currentMessage, elapsed.Truncate(time.Second))
	
	pt.printProgressBar()
}

// printProgressBar 打印进度条
func (pt *ProgressTracker) printProgressBar() {
	barWidth := 30
	filled := int(pt.percentage / 100 * float64(barWidth))
	
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	fmt.Printf(" [%s] %.1f%%", bar, pt.percentage)
}

// GetCurrentStatus 获取当前状态
func (pt *ProgressTracker) GetCurrentStatus() map[string]interface{} {
	pt.mutex.RLock()
	defer pt.mutex.RUnlock()

	return map[string]interface{}{
		"task_name":        pt.taskName,
		"total_steps":      pt.totalSteps,
		"current_step":     pt.currentStep,
		"current_message":  pt.currentMessage,
		"percentage":       pt.percentage,
		"start_time":       pt.startTime,
		"last_update_time": pt.lastUpdateTime,
		"is_running":       pt.isRunning,
		"elapsed_time":     time.Since(pt.startTime),
	}
}

// GetProgress 获取进度百分比
func (pt *ProgressTracker) GetProgress() float64 {
	pt.mutex.RLock()
	defer pt.mutex.RUnlock()
	return pt.percentage
}

// GetCurrentStep 获取当前步骤
func (pt *ProgressTracker) GetCurrentStep() int {
	pt.mutex.RLock()
	defer pt.mutex.RUnlock()
	return pt.currentStep
}

// GetTotalSteps 获取总步骤数
func (pt *ProgressTracker) GetTotalSteps() int {
	pt.mutex.RLock()
	defer pt.mutex.RUnlock()
	return pt.totalSteps
}

// GetCurrentMessage 获取当前消息
func (pt *ProgressTracker) GetCurrentMessage() string {
	pt.mutex.RLock()
	defer pt.mutex.RUnlock()
	return pt.currentMessage
}

// IsRunning 检查是否正在运行
func (pt *ProgressTracker) IsRunning() bool {
	pt.mutex.RLock()
	defer pt.mutex.RUnlock()
	return pt.isRunning
}

// GetElapsedTime 获取已用时间
func (pt *ProgressTracker) GetElapsedTime() time.Duration {
	pt.mutex.RLock()
	defer pt.mutex.RUnlock()
	
	if pt.startTime.IsZero() {
		return 0
	}
	
	return time.Since(pt.startTime)
}

// GetEstimatedTimeRemaining 获取预计剩余时间
func (pt *ProgressTracker) GetEstimatedTimeRemaining() time.Duration {
	pt.mutex.RLock()
	defer pt.mutex.RUnlock()

	if pt.percentage <= 0 || pt.startTime.IsZero() {
		return 0
	}

	elapsed := time.Since(pt.startTime)
	totalEstimated := time.Duration(float64(elapsed) / pt.percentage * 100)
	remaining := totalEstimated - elapsed

	if remaining < 0 {
		return 0
	}

	return remaining
}

// SetMessage 设置当前消息
func (pt *ProgressTracker) SetMessage(message string) {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()

	pt.currentMessage = message
	pt.lastUpdateTime = time.Now()
}

// AddStep 增加一个步骤
func (pt *ProgressTracker) AddStep(message string) {
	pt.UpdateStep(pt.GetCurrentStep()+1, message)
}

// Complete 标记为完成
func (pt *ProgressTracker) Complete(message string) {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()

	pt.currentStep = pt.totalSteps
	pt.percentage = 100
	pt.currentMessage = message
	pt.lastUpdateTime = time.Now()

	fmt.Printf("\r✅ %s - %s [████████████████████████████████] 100.0%%\n", 
		pt.taskName, message)
}
