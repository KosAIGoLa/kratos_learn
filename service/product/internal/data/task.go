package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"product/internal/biz"
)

// Task 任务数据模型
type Task struct {
	ID         uint32  `gorm:"primarykey"`
	Name       string  `gorm:"type:varchar(100);not null"`
	Code       string  `gorm:"uniqueIndex:idx_code;type:varchar(50);not null"`
	WorkPoints float64 `gorm:"type:decimal(10,2);default:0.00"`
	Enabled    int8    `gorm:"index:idx_enabled;default:1"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// UserTask 用户任务数据模型
type UserTask struct {
	ID              uint32  `gorm:"primarykey"`
	UserID          uint32  `gorm:"index:idx_user_id;not null"`
	TaskID          uint32  `gorm:"index:idx_task_id;not null"`
	CompletedTimes  uint32  `gorm:"default:0"`
	TotalReward     float64 `gorm:"type:decimal(10,2);default:0.00"`
	Status          int8    `gorm:"index:idx_status;default:1"`
	LastCompletedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type taskRepo struct {
	data *Data
	log  *log.Helper
}

// NewTaskRepo 创建任务仓库
func NewTaskRepo(data *Data, logger log.Logger) biz.TaskRepo {
	return &taskRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *taskRepo) ListTasks(ctx context.Context, enabled int32) ([]*biz.Task, error) {
	var tasks []Task
	query := r.data.db.Model(&Task{})
	if enabled >= 0 {
		query = query.Where("enabled = ?", enabled)
	}
	if err := query.Find(&tasks).Error; err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	var bizTasks []*biz.Task
	for _, t := range tasks {
		bizTasks = append(bizTasks, &biz.Task{
			ID:         t.ID,
			Name:       t.Name,
			Code:       t.Code,
			WorkPoints: t.WorkPoints,
			Enabled:    int32(t.Enabled),
			CreatedAt:  t.CreatedAt,
		})
	}
	return bizTasks, nil
}

func (r *taskRepo) GetTask(ctx context.Context, id uint32) (*biz.Task, error) {
	var task Task
	if err := r.data.db.First(&task, id).Error; err != nil {
		return nil, status.Errorf(codes.NotFound, "任务不存在")
	}
	return &biz.Task{
		ID:         task.ID,
		Name:       task.Name,
		Code:       task.Code,
		WorkPoints: task.WorkPoints,
		Enabled:    int32(task.Enabled),
	}, nil
}

func (r *taskRepo) CompleteTask(ctx context.Context, userTask *biz.UserTask) (*biz.UserTask, error) {
	ut := UserTask{
		UserID:         userTask.UserID,
		TaskID:         userTask.TaskID,
		CompletedTimes: 1,
		TotalReward:    userTask.TotalReward,
		Status:         1,
	}
	now := time.Now()
	ut.LastCompletedAt = &now

	if err := r.data.db.Create(&ut).Error; err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &biz.UserTask{
		ID:              ut.ID,
		UserID:          ut.UserID,
		TaskID:          ut.TaskID,
		CompletedTimes:  ut.CompletedTimes,
		TotalReward:     ut.TotalReward,
		Status:          int32(ut.Status),
		LastCompletedAt: ut.LastCompletedAt,
	}, nil
}
