package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"system/internal/biz"
)

// systemLogRepo 系统日志存储实现
type systemLogRepo struct {
	data *Data
	log  *log.Helper
}

// NewSystemLogRepo 创建系统日志存储
func NewSystemLogRepo(data *Data, logger log.Logger) biz.SystemLogRepo {
	return &systemLogRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// getCollection 获取系统日志集合
func (r *systemLogRepo) getCollection() *mongo.Collection {
	if r.data.mdb == nil {
		return nil
	}
	return r.data.mdb.Collection("system_logs")
}

// ListSystemLogs 查询系统日志列表
func (r *systemLogRepo) ListSystemLogs(ctx context.Context, level, module, operatorID, startTime, endTime string, page, pageSize int32) ([]*biz.SystemLog, int32, error) {
	coll := r.getCollection()
	if coll == nil {
		return nil, 0, status.Errorf(codes.Internal, "mongodb not connected")
	}

	// 构建过滤条件
	filter := bson.M{}
	if level != "" {
		filter["level"] = level
	}
	if module != "" {
		filter["module"] = module
	}
	if operatorID != "" {
		filter["operator_id"] = operatorID
	}
	if startTime != "" || endTime != "" {
		dateFilter := bson.M{}
		if startTime != "" {
			start, err := time.Parse(time.RFC3339, startTime)
			if err == nil {
				dateFilter["$gte"] = start
			}
		}
		if endTime != "" {
			end, err := time.Parse(time.RFC3339, endTime)
			if err == nil {
				dateFilter["$lte"] = end
			}
		}
		if len(dateFilter) > 0 {
			filter["created_at"] = dateFilter
		}
	}

	// 计算总数
	total, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		r.log.Errorf("count system logs failed: %v", err)
		return nil, 0, status.Errorf(codes.Internal, "%s", err.Error())
	}

	// 查询数据
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize))

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		r.log.Errorf("find system logs failed: %v", err)
		return nil, 0, status.Errorf(codes.Internal, "%s", err.Error())
	}
	defer func() { _ = cursor.Close(ctx) }()

	var logs []*biz.SystemLog
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		logs = append(logs, r.docToSystemLog(doc))
	}

	return logs, int32(total), nil
}

// GetSystemLog 获取单条系统日志
func (r *systemLogRepo) GetSystemLog(ctx context.Context, id string) (*biz.SystemLog, error) {
	coll := r.getCollection()
	if coll == nil {
		return nil, status.Errorf(codes.Internal, "mongodb not connected")
	}

	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id format")
	}

	var doc bson.M
	err = coll.FindOne(ctx, bson.M{"_id": objectID}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, status.Errorf(codes.NotFound, "system log not found")
		}
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	return r.docToSystemLog(doc), nil
}

// CreateSystemLog 创建系统日志
func (r *systemLogRepo) CreateSystemLog(ctx context.Context, log *biz.SystemLog) error {
	coll := r.getCollection()
	if coll == nil {
		return status.Errorf(codes.Internal, "mongodb not connected")
	}

	doc := bson.M{
		"level":         log.Level,
		"module":        log.Module,
		"action":        log.Action,
		"message":       log.Message,
		"operator_id":   log.OperatorID,
		"operator_name": log.OperatorName,
		"ip_address":    log.IPAddress,
		"user_agent":    log.UserAgent,
		"metadata":      log.Metadata,
		"created_at":    log.CreatedAt,
	}

	_, err := coll.InsertOne(ctx, doc)
	if err != nil {
		r.log.Errorf("create system log failed: %v", err)
		return status.Errorf(codes.Internal, "%s", err.Error())
	}

	return nil
}

// docToSystemLog 将 MongoDB 文档转换为实体
func (r *systemLogRepo) docToSystemLog(doc bson.M) *biz.SystemLog {
	log := &biz.SystemLog{
		Level:        getStringValue(doc, "level"),
		Module:       getStringValue(doc, "module"),
		Action:       getStringValue(doc, "action"),
		Message:      getStringValue(doc, "message"),
		OperatorID:   getStringValue(doc, "operator_id"),
		OperatorName: getStringValue(doc, "operator_name"),
		IPAddress:    getStringValue(doc, "ip_address"),
		UserAgent:    getStringValue(doc, "user_agent"),
		Metadata:     make(map[string]string),
	}

	// 设置 ID
	if id, ok := doc["_id"]; ok {
		if oid, ok := id.(bson.ObjectID); ok {
			log.ID = oid.Hex()
		}
	}

	// 设置时间
	if createdAt, ok := doc["created_at"]; ok {
		if t, ok := createdAt.(time.Time); ok {
			log.CreatedAt = t
		}
	}

	// 设置元数据
	if meta, ok := doc["metadata"]; ok {
		if m, ok := meta.(bson.M); ok {
			for k, v := range m {
				if s, ok := v.(string); ok {
					log.Metadata[k] = s
				}
			}
		}
	}

	return log
}

// userLogRepo 用户日志存储实现
type userLogRepo struct {
	data *Data
	log  *log.Helper
}

// NewUserLogRepo 创建用户日志存储
func NewUserLogRepo(data *Data, logger log.Logger) biz.UserLogRepo {
	return &userLogRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// getCollection 获取用户日志集合
func (r *userLogRepo) getCollection() *mongo.Collection {
	if r.data.mdb == nil {
		return nil
	}
	return r.data.mdb.Collection("user_logs")
}

// ListUserLogs 查询用户日志列表
func (r *userLogRepo) ListUserLogs(ctx context.Context, userID uint32, action, module, startTime, endTime string, page, pageSize int32) ([]*biz.UserLog, int32, error) {
	coll := r.getCollection()
	if coll == nil {
		return nil, 0, status.Errorf(codes.Internal, "mongodb not connected")
	}

	// 构建过滤条件
	filter := bson.M{}
	if userID > 0 {
		filter["user_id"] = userID
	}
	if action != "" {
		filter["action"] = action
	}
	if module != "" {
		filter["module"] = module
	}
	if startTime != "" || endTime != "" {
		dateFilter := bson.M{}
		if startTime != "" {
			start, err := time.Parse(time.RFC3339, startTime)
			if err == nil {
				dateFilter["$gte"] = start
			}
		}
		if endTime != "" {
			end, err := time.Parse(time.RFC3339, endTime)
			if err == nil {
				dateFilter["$lte"] = end
			}
		}
		if len(dateFilter) > 0 {
			filter["created_at"] = dateFilter
		}
	}

	// 计算总数
	total, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		r.log.Errorf("count user logs failed: %v", err)
		return nil, 0, status.Errorf(codes.Internal, "%s", err.Error())
	}

	// 查询数据
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize))

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		r.log.Errorf("find user logs failed: %v", err)
		return nil, 0, status.Errorf(codes.Internal, "%s", err.Error())
	}
	defer func() { _ = cursor.Close(ctx) }()

	var logs []*biz.UserLog
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		logs = append(logs, r.docToUserLog(doc))
	}

	return logs, int32(total), nil
}

// GetUserLog 获取单条用户日志
func (r *userLogRepo) GetUserLog(ctx context.Context, id string) (*biz.UserLog, error) {
	coll := r.getCollection()
	if coll == nil {
		return nil, status.Errorf(codes.Internal, "mongodb not connected")
	}

	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id format")
	}

	var doc bson.M
	err = coll.FindOne(ctx, bson.M{"_id": objectID}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, status.Errorf(codes.NotFound, "user log not found")
		}
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	return r.docToUserLog(doc), nil
}

// CreateUserLog 创建用户日志
func (r *userLogRepo) CreateUserLog(ctx context.Context, log *biz.UserLog) error {
	coll := r.getCollection()
	if coll == nil {
		return status.Errorf(codes.Internal, "mongodb not connected")
	}

	doc := bson.M{
		"user_id":     log.UserID,
		"username":    log.Username,
		"action":      log.Action,
		"module":      log.Module,
		"description": log.Description,
		"ip_address":  log.IPAddress,
		"device_info": log.DeviceInfo,
		"metadata":    log.Metadata,
		"created_at":  log.CreatedAt,
	}

	_, err := coll.InsertOne(ctx, doc)
	if err != nil {
		r.log.Errorf("create user log failed: %v", err)
		return status.Errorf(codes.Internal, "%s", err.Error())
	}

	return nil
}

// docToUserLog 将 MongoDB 文档转换为实体
func (r *userLogRepo) docToUserLog(doc bson.M) *biz.UserLog {
	log := &biz.UserLog{
		Username:    getStringValue(doc, "username"),
		Action:      getStringValue(doc, "action"),
		Module:      getStringValue(doc, "module"),
		Description: getStringValue(doc, "description"),
		IPAddress:   getStringValue(doc, "ip_address"),
		DeviceInfo:  getStringValue(doc, "device_info"),
		Metadata:    make(map[string]string),
	}

	// 设置 ID
	if id, ok := doc["_id"]; ok {
		if oid, ok := id.(bson.ObjectID); ok {
			log.ID = oid.Hex()
		}
	}

	// 设置 UserID
	if uid, ok := doc["user_id"]; ok {
		switch v := uid.(type) {
		case int32:
			log.UserID = uint32(v)
		case int64:
			log.UserID = uint32(v)
		case float64:
			log.UserID = uint32(v)
		}
	}

	// 设置时间
	if createdAt, ok := doc["created_at"]; ok {
		if t, ok := createdAt.(time.Time); ok {
			log.CreatedAt = t
		}
	}

	// 设置元数据
	if meta, ok := doc["metadata"]; ok {
		if m, ok := meta.(bson.M); ok {
			for k, v := range m {
				if s, ok := v.(string); ok {
					log.Metadata[k] = s
				}
			}
		}
	}

	return log
}

// getStringValue 从 bson.M 中获取字符串值
func getStringValue(doc bson.M, key string) string {
	if val, ok := doc[key]; ok {
		switch v := val.(type) {
		case string:
			return v
		case bson.ObjectID:
			return v.Hex()
		default:
			return ""
		}
	}
	return ""
}
