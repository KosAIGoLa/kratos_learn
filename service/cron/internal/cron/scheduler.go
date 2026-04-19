package cron

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/robfig/cron/v3"
)

// Scheduler 定时任务调度器
type Scheduler struct {
	c     *cron.Cron
	log   *log.Helper
	tasks map[string]cron.EntryID
	mu    sync.RWMutex
}

// NewScheduler 创建新的调度器
func NewScheduler(logger log.Logger) *Scheduler {
	return &Scheduler{
		c: cron.New(
			cron.WithSeconds(), // 支持秒级
		),
		log:   log.NewHelper(logger),
		tasks: make(map[string]cron.EntryID),
	}
}

// AddTask 添加定时任务
// spec: cron 表达式，例如 "0 0 * * * *" (每秒), "0 0 2 * * *" (每天2点)
// name: 任务名称
// fn: 任务函数
func (s *Scheduler) AddTask(name string, spec string, fn func()) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 如果任务已存在，先移除
	if id, exists := s.tasks[name]; exists {
		s.c.Remove(id)
		s.log.Infof("移除已存在的任务: %s", name)
	}

	entryID, err := s.c.AddFunc(spec, fn)
	if err != nil {
		return fmt.Errorf("添加任务失败 %s: %w", name, err)
	}

	s.tasks[name] = entryID
	s.log.Infof("添加定时任务成功: %s, spec: %s, entryID: %d", name, spec, entryID)
	return nil
}

// RemoveTask 移除定时任务
func (s *Scheduler) RemoveTask(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id, exists := s.tasks[name]; exists {
		s.c.Remove(id)
		delete(s.tasks, name)
		s.log.Infof("移除任务: %s", name)
	}
}

// Start 启动调度器
func (s *Scheduler) Start(ctx context.Context) error {
	s.c.Start()
	s.log.Info("定时任务调度器已启动")
	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop(ctx context.Context) error {
	s.log.Info("正在停止定时任务调度器...")
	<-s.c.Stop().Done()
	s.log.Info("定时任务调度器已停止")
	return nil
}

// ListTasks 列出所有任务
func (s *Scheduler) ListTasks() []TaskInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var tasks []TaskInfo
	for name, entryID := range s.tasks {
		entry := s.c.Entry(entryID)
		tasks = append(tasks, TaskInfo{
			Name:    name,
			NextRun: entry.Next.Format(time.RFC3339),
			PrevRun: entry.Prev.Format(time.RFC3339),
		})
	}
	return tasks
}

// TaskInfo 任务信息
type TaskInfo struct {
	Name    string `json:"name"`
	NextRun string `json:"next_run"`
	PrevRun string `json:"prev_run"`
}
