package service

import (
	"context"
	"fmt"

	v1 "cron/api/cron/v1"
	"cron/internal/cron"

	"github.com/go-kratos/kratos/v2/log"
)

// CronService 定时任务管理服务
type CronService struct {
	v1.UnimplementedCronServer

	taskManager *cron.TaskManager
	log         *log.Helper
}

// NewCronService 创建定时任务管理服务
func NewCronService(tm *cron.TaskManager, logger log.Logger) *CronService {
	return &CronService{
		taskManager: tm,
		log:         log.NewHelper(logger),
	}
}

// ListTasks 列出所有定时任务
func (s *CronService) ListTasks(ctx context.Context, req *v1.ListTasksRequest) (*v1.ListTasksReply, error) {
	tasks := s.taskManager.ListTasks()

	var taskList []*v1.TaskInfo
	for _, t := range tasks {
		taskList = append(taskList, &v1.TaskInfo{
			Name:    t.Name,
			NextRun: t.NextRun,
			PrevRun: t.PrevRun,
		})
	}

	return &v1.ListTasksReply{
		Tasks: taskList,
	}, nil
}

// AddTask 添加定时任务
func (s *CronService) AddTask(ctx context.Context, req *v1.AddTaskRequest) (*v1.AddTaskReply, error) {
	if req.Name == "" || req.Spec == "" {
		return nil, fmt.Errorf("任务名称和 cron 表达式不能为空")
	}

	// 这里可以添加更多复杂的任务逻辑
	// 简单示例：添加一个打印日志的任务
	task := func() {
		s.log.Infof("执行自定义任务: %s", req.Name)
	}

	if err := s.taskManager.AddCustomTask(req.Name, req.Spec, task); err != nil {
		return nil, err
	}

	return &v1.AddTaskReply{
		Success: true,
		Message: fmt.Sprintf("任务 %s 添加成功", req.Name),
	}, nil
}

// RemoveTask 移除定时任务
func (s *CronService) RemoveTask(ctx context.Context, req *v1.RemoveTaskRequest) (*v1.RemoveTaskReply, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("任务名称不能为空")
	}

	s.taskManager.RemoveTask(req.Name)

	return &v1.RemoveTaskReply{
		Success: true,
		Message: fmt.Sprintf("任务 %s 已移除", req.Name),
	}, nil
}
