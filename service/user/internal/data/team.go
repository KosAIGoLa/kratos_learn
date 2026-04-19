package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"user/internal/biz"
)

// TeamRelation 团队关系数据模型
type TeamRelation struct {
	ID        uint32  `gorm:"primarykey"`
	UserID    uint32  `gorm:"uniqueIndex:uk_user_id;not null"`
	ParentID  *uint32 `gorm:"index:idx_parent_id"`
	Path      string  `gorm:"type:varchar(500);not null"`
	Level     uint32  `gorm:"index:idx_level;not null"`
	CreatedAt time.Time
}

type teamRepo struct {
	data *Data
	log  *log.Helper
}

// NewTeamRepo 创建团队关系仓库
func NewTeamRepo(data *Data, logger log.Logger) biz.TeamRepo {
	return &teamRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *teamRepo) CreateTeamRelation(ctx context.Context, t *biz.TeamRelation) (*biz.TeamRelation, error) {
	relation := TeamRelation{
		UserID:   t.UserID,
		ParentID: t.ParentID,
		Path:     t.Path,
		Level:    t.Level,
	}
	if err := r.data.db.Create(&relation).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	return &biz.TeamRelation{
		ID:        relation.ID,
		UserID:    relation.UserID,
		ParentID:  relation.ParentID,
		Path:      relation.Path,
		Level:     relation.Level,
		CreatedAt: relation.CreatedAt,
	}, nil
}

func (r *teamRepo) GetTeamRelationByUserID(ctx context.Context, userID uint32) (*biz.TeamRelation, error) {
	var relation TeamRelation
	if err := r.data.db.Where("user_id = ?", userID).First(&relation).Error; err != nil {
		return nil, err
	}
	return &biz.TeamRelation{
		ID:        relation.ID,
		UserID:    relation.UserID,
		ParentID:  relation.ParentID,
		Path:      relation.Path,
		Level:     relation.Level,
		CreatedAt: relation.CreatedAt,
	}, nil
}

func (r *teamRepo) GetTeamMembers(ctx context.Context, parentID uint32, level int32) ([]*biz.User, error) {
	var relations []TeamRelation
	query := r.data.db.Where("parent_id = ?", parentID)
	if level >= 0 {
		query = query.Where("level = ?", level)
	}
	if err := query.Find(&relations).Error; err != nil {
		return nil, err
	}

	var userIDs []uint32
	for _, rel := range relations {
		userIDs = append(userIDs, rel.UserID)
	}

	if len(userIDs) == 0 {
		return []*biz.User{}, nil
	}

	var users []User
	r.data.db.Where("id IN ?", userIDs).Find(&users)

	var bizUsers []*biz.User
	for _, u := range users {
		bizUsers = append(bizUsers, &biz.User{
			ID:         u.ID,
			Username:   u.Username,
			InviteCode: u.InviteCode,
			Phone:      u.Phone,
			Name:       u.Name,
			Status:     int32(u.Status),
		})
	}
	return bizUsers, nil
}

func (r *teamRepo) GetAllDescendants(ctx context.Context, userID uint32) ([]*biz.User, error) {
	var relations []TeamRelation
	if err := r.data.db.Where("path LIKE ?", "%"+string(rune(userID))+"%").Find(&relations).Error; err != nil {
		return nil, err
	}
	return []*biz.User{}, nil
}
