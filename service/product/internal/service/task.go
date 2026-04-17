package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	v1 "product/api/product/v1"
	"product/internal/biz"
)

// TaskService 任务服务
type TaskService struct {
	v1.UnimplementedProductServer
	uc  *biz.TaskUsecase
	log *log.Helper
}

// NewTaskService 创建任务服务
func NewTaskService(uc *biz.TaskUsecase, logger log.Logger) *TaskService {
	return &TaskService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// ListTasks 获取任务列表
func (s *TaskService) ListTasks(ctx context.Context, req *v1.ListTasksRequest) (*v1.ListTasksResponse, error) {
	tasks, err := s.uc.ListTasks(ctx, req.Enabled)
	if err != nil {
		return nil, err
	}

	var protoTasks []*v1.TaskInfo
	for _, t := range tasks {
		protoTasks = append(protoTasks, &v1.TaskInfo{
			Id:         t.ID,
			Name:       t.Name,
			Code:       t.Code,
			WorkPoints: t.WorkPoints,
			Enabled:    t.Enabled,
		})
	}

	return &v1.ListTasksResponse{
		Tasks: protoTasks,
	}, nil
}

// GetTask 获取任务详情
func (s *TaskService) GetTask(ctx context.Context, req *v1.GetTaskRequest) (*v1.TaskInfo, error) {
	task, err := s.uc.GetTask(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.TaskInfo{
		Id:         task.ID,
		Name:       task.Name,
		Code:       task.Code,
		WorkPoints: task.WorkPoints,
		Enabled:    task.Enabled,
	}, nil
}

// CompleteTask 完成任务
func (s *TaskService) CompleteTask(ctx context.Context, req *v1.CompleteTaskRequest) (*v1.UserTaskInfo, error) {
	userTask, err := s.uc.CompleteTask(ctx, &biz.UserTask{
		UserID: req.UserId,
		TaskID: req.TaskId,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UserTaskInfo{
		Id:             userTask.ID,
		UserId:         userTask.UserID,
		TaskId:         userTask.TaskID,
		CompletedTimes: userTask.CompletedTimes,
		TotalReward:    userTask.TotalReward,
		Status:         userTask.Status,
	}, nil
}
