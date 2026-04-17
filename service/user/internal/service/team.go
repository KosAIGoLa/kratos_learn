package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	v1 "user/api/user/v1"
	"user/internal/biz"
)

// TeamService 团队服务
type TeamService struct {
	v1.UnimplementedUserServer
	uc  *biz.TeamUsecase
	log *log.Helper
}

// NewTeamService 创建团队服务
func NewTeamService(uc *biz.TeamUsecase, logger log.Logger) *TeamService {
	return &TeamService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// GetTeamRelation 获取团队关系
func (s *TeamService) GetTeamRelation(ctx context.Context, req *v1.GetTeamRequest) (*v1.TeamRelationInfo, error) {
	relation, err := s.uc.GetTeamRelationByUserID(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	return &v1.TeamRelationInfo{
		Id:     relation.ID,
		UserId: relation.UserID,
		ParentId: func() uint32 {
			if relation.ParentID != nil {
				return *relation.ParentID
			}
			return 0
		}(),
		Path:  relation.Path,
		Level: relation.Level,
	}, nil
}

// GetTeamMembers 获取团队成员
func (s *TeamService) GetTeamMembers(ctx context.Context, req *v1.GetTeamMembersRequest) (*v1.TeamMembersResponse, error) {
	members, err := s.uc.GetTeamMembers(ctx, req.UserId, req.Level)
	if err != nil {
		return nil, err
	}

	var protoMembers []*v1.UserInfo
	for _, m := range members {
		protoMembers = append(protoMembers, &v1.UserInfo{
			Id:         m.ID,
			Username:   m.Username,
			InviteCode: m.InviteCode,
			Phone:      m.Phone,
			Name:       m.Name,
			Status:     m.Status,
		})
	}

	return &v1.TeamMembersResponse{
		Members: protoMembers,
		Total:   uint32(len(protoMembers)),
	}, nil
}
