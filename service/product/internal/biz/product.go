package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// Product 产品领域模型
type Product struct {
	ID               uint32
	MachineID        *uint32
	Name             string
	Price            float64
	Description      string
	Type             string
	Cycle            uint32
	ProductivityRate float64
	Status           int32
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// MiningMachine 矿机领域模型
type MiningMachine struct {
	ID                  uint32
	Name                string
	Model               string
	Algorithm           string
	Hashrate            float64
	HashrateUnit        string
	PowerConsumption    uint32
	DurationDays        uint32
	Price               float64
	DailyRewardEstimate float64
	Stock               uint32
	Status              int32
	CreatedAt           time.Time
}

// Task 任务领域模型
type Task struct {
	ID         uint32
	Name       string
	Code       string
	WorkPoints float64
	Enabled    int32
	CreatedAt  time.Time
}

// UserTask 用户任务领域模型
type UserTask struct {
	ID              uint32
	UserID          uint32
	TaskID          uint32
	CompletedTimes  uint32
	TotalReward     float64
	Status          int32
	LastCompletedAt *time.Time
	CreatedAt       time.Time
}

// ProductRepo 产品存储接口
type ProductRepo interface {
	ListProducts(ctx context.Context, typ string, status int32, page, pageSize uint32) ([]*Product, uint32, error)
	GetProduct(ctx context.Context, id uint32) (*Product, error)
	CreateProduct(ctx context.Context, p *Product) (*Product, error)
	UpdateProduct(ctx context.Context, p *Product) (*Product, error)
	DeleteProduct(ctx context.Context, id uint32) error
}

// MiningMachineRepo 矿机存储接口
type MiningMachineRepo interface {
	ListMachines(ctx context.Context, algorithm string, status int32) ([]*MiningMachine, error)
	GetMachine(ctx context.Context, id uint32) (*MiningMachine, error)
}

// TaskRepo 任务存储接口
type TaskRepo interface {
	ListTasks(ctx context.Context, enabled int32) ([]*Task, error)
	GetTask(ctx context.Context, id uint32) (*Task, error)
	CompleteTask(ctx context.Context, userTask *UserTask) (*UserTask, error)
}

// ProductUsecase 产品用例
type ProductUsecase struct {
	repo ProductRepo
	log  *log.Helper
}

// NewProductUsecase 创建产品用例
func NewProductUsecase(repo ProductRepo, logger log.Logger) *ProductUsecase {
	return &ProductUsecase{repo: repo, log: log.NewHelper(logger)}
}

// ListProducts 获取产品列表
func (uc *ProductUsecase) ListProducts(ctx context.Context, typ string, status int32, page, pageSize uint32) ([]*Product, uint32, error) {
	return uc.repo.ListProducts(ctx, typ, status, page, pageSize)
}

// GetProduct 获取产品
func (uc *ProductUsecase) GetProduct(ctx context.Context, id uint32) (*Product, error) {
	return uc.repo.GetProduct(ctx, id)
}

// CreateProduct 创建产品
func (uc *ProductUsecase) CreateProduct(ctx context.Context, p *Product) (*Product, error) {
	return uc.repo.CreateProduct(ctx, p)
}

// UpdateProduct 更新产品
func (uc *ProductUsecase) UpdateProduct(ctx context.Context, p *Product) (*Product, error) {
	return uc.repo.UpdateProduct(ctx, p)
}

// DeleteProduct 删除产品
func (uc *ProductUsecase) DeleteProduct(ctx context.Context, id uint32) error {
	return uc.repo.DeleteProduct(ctx, id)
}

// TaskUsecase 任务用例
type TaskUsecase struct {
	repo TaskRepo
	log  *log.Helper
}

// NewTaskUsecase 创建任务用例
func NewTaskUsecase(repo TaskRepo, logger log.Logger) *TaskUsecase {
	return &TaskUsecase{repo: repo, log: log.NewHelper(logger)}
}

// ListTasks 获取任务列表
func (uc *TaskUsecase) ListTasks(ctx context.Context, enabled int32) ([]*Task, error) {
	return uc.repo.ListTasks(ctx, enabled)
}

// GetTask 获取任务
func (uc *TaskUsecase) GetTask(ctx context.Context, id uint32) (*Task, error) {
	return uc.repo.GetTask(ctx, id)
}

// CompleteTask 完成任务
func (uc *TaskUsecase) CompleteTask(ctx context.Context, userTask *UserTask) (*UserTask, error) {
	return uc.repo.CompleteTask(ctx, userTask)
}
