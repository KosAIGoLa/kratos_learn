package data

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"sms/internal/biz"
)

// SmsLogModel 短信发送记录数据库模型
type SmsLogModel struct {
	ID                string `gorm:"primary_key;size:64"`
	Phone             string `gorm:"index;size:20"`
	TemplateCode      string `gorm:"size:64"`
	TemplateParams    string `gorm:"type:text"`
	Content           string `gorm:"type:text"`
	ProviderID        string `gorm:"index;size:64"`
	ProviderName      string `gorm:"size:128"`
	Status            string `gorm:"index;size:20"` // pending, sent, failed, delivered
	ProviderMessageID string `gorm:"size:128"`
	ErrorMessage      string `gorm:"type:text"`
	RetryCount        int32
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// TableName 表名
func (SmsLogModel) TableName() string {
	return "sms_logs"
}

// smsRepo 短信记录存储实现
type smsRepo struct {
	data *Data
	log  *log.Helper
}

// NewSmsRepo 创建短信记录存储实例
func NewSmsRepo(data *Data, logger log.Logger) biz.SmsRepo {
	return &smsRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// SaveLog 保存短信发送记录
func (r *smsRepo) SaveLog(ctx context.Context, log *biz.SmsLog) error {
	// 序列化模板参数
	paramsJson, _ := json.Marshal(log.TemplateParams)

	model := &SmsLogModel{
		ID:                log.ID,
		Phone:             log.Phone,
		TemplateCode:      log.TemplateCode,
		TemplateParams:    string(paramsJson),
		Content:           log.Content,
		ProviderID:        log.ProviderID,
		ProviderName:      log.ProviderName,
		Status:            log.Status,
		ProviderMessageID: log.ProviderMessageID,
		ErrorMessage:      log.ErrorMessage,
		RetryCount:        log.RetryCount,
		CreatedAt:         log.CreatedAt,
		UpdatedAt:         log.UpdatedAt,
	}

	// 使用 Upsert 保存记录
	return r.data.db.WithContext(ctx).Save(model).Error
}

// GetLog 获取短信发送记录
func (r *smsRepo) GetLog(ctx context.Context, id string) (*biz.SmsLog, error) {
	var model SmsLogModel
	if err := r.data.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.toBizLog(&model), nil
}

// ListLogs 查询短信发送记录列表
func (r *smsRepo) ListLogs(ctx context.Context, phone, status, providerId string, page, pageSize int32) ([]*biz.SmsLog, int32, error) {
	query := r.data.db.WithContext(ctx).Model(&SmsLogModel{})

	// 添加筛选条件
	if phone != "" {
		query = query.Where("phone = ?", phone)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if providerId != "" {
		query = query.Where("provider_id = ?", providerId)
	}

	// 查询总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	var models []SmsLogModel
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&models).Error; err != nil {
		return nil, 0, err
	}

	// 转换为业务对象
	logs := make([]*biz.SmsLog, 0, len(models))
	for _, m := range models {
		logs = append(logs, r.toBizLog(&m))
	}

	return logs, int32(total), nil
}

// toBizLog 转换为业务对象
func (r *smsRepo) toBizLog(model *SmsLogModel) *biz.SmsLog {
	// 反序列化模板参数
	var params map[string]string
	if err := json.Unmarshal([]byte(model.TemplateParams), &params); err != nil {
		r.log.Warnf("failed to unmarshal template params for sms log %s: %v", model.ID, err)
	}

	return &biz.SmsLog{
		ID:                model.ID,
		Phone:             model.Phone,
		TemplateCode:      model.TemplateCode,
		TemplateParams:    params,
		Content:           model.Content,
		ProviderID:        model.ProviderID,
		ProviderName:      model.ProviderName,
		Status:            model.Status,
		ProviderMessageID: model.ProviderMessageID,
		ErrorMessage:      model.ErrorMessage,
		RetryCount:        model.RetryCount,
		CreatedAt:         model.CreatedAt,
		UpdatedAt:         model.UpdatedAt,
	}
}
