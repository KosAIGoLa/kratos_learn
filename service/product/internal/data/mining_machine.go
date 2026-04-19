package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"product/internal/biz"
)

// MiningMachine 矿机数据模型
type MiningMachine struct {
	ID                  uint32  `gorm:"primarykey"`
	Name                string  `gorm:"type:varchar(100);not null"`
	Model               string  `gorm:"type:varchar(50);not null"`
	Algorithm           string  `gorm:"index:idx_algorithm;type:varchar(30);not null"`
	Hashrate            float64 `gorm:"type:decimal(15,2);not null"`
	HashrateUnit        string  `gorm:"type:varchar(10);not null"`
	PowerConsumption    uint32
	DurationDays        uint32  `gorm:"not null"`
	Price               float64 `gorm:"type:decimal(15,2);not null"`
	DailyRewardEstimate float64 `gorm:"type:decimal(15,6)"`
	Stock               uint32  `gorm:"default:0"`
	Status              int8    `gorm:"index:idx_status;default:1"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type miningMachineRepo struct {
	data *Data
	log  *log.Helper
}

// NewMiningMachineRepo 创建矿机仓库
func NewMiningMachineRepo(data *Data, logger log.Logger) biz.MiningMachineRepo {
	return &miningMachineRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *miningMachineRepo) ListMachines(ctx context.Context, algorithm string, statusFilter int32) ([]*biz.MiningMachine, error) {
	var machines []MiningMachine
	query := r.data.db.Model(&MiningMachine{})
	if algorithm != "" {
		query = query.Where("algorithm = ?", algorithm)
	}
	if statusFilter >= 0 {
		query = query.Where("status = ?", statusFilter)
	}
	if err := query.Find(&machines).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	var bizMachines []*biz.MiningMachine
	for _, m := range machines {
		bizMachines = append(bizMachines, r.toBizMachine(&m))
	}
	return bizMachines, nil
}

func (r *miningMachineRepo) GetMachine(ctx context.Context, id uint32) (*biz.MiningMachine, error) {
	var machine MiningMachine
	if err := r.data.db.First(&machine, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "矿机不存在")
		}
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	return r.toBizMachine(&machine), nil
}

func (r *miningMachineRepo) toBizMachine(m *MiningMachine) *biz.MiningMachine {
	return &biz.MiningMachine{
		ID:                  m.ID,
		Name:                m.Name,
		Model:               m.Model,
		Algorithm:           m.Algorithm,
		Hashrate:            m.Hashrate,
		HashrateUnit:        m.HashrateUnit,
		PowerConsumption:    m.PowerConsumption,
		DurationDays:        m.DurationDays,
		Price:               m.Price,
		DailyRewardEstimate: m.DailyRewardEstimate,
		Stock:               m.Stock,
		Status:              int32(m.Status),
	}
}
