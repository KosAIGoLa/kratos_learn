---
trigger: manual
description: Kratos API 定义规范 (Proto)
globs: ["**/api/**/*.proto"]
---

# Kratos API 定义规范

## 1. Proto 文件结构

```protobuf
syntax = "proto3";

package {service}.v1;

option go_package = "{service}/api/{service}/v1;v1";

import "google/protobuf/timestamp.proto";
import "validate/validate.proto";

// 服务定义
service {Service} {
  rpc Create{Entity}(Create{Entity}Request) returns (Create{Entity}Response);
  rpc Get{Entity}(Get{Entity}Request) returns ({Entity}Info);
  rpc Update{Entity}(Update{Entity}Request) returns (Update{Entity}Response);
  rpc Delete{Entity}(Delete{Entity}Request) returns (Delete{Entity}Response);
  rpc List{Entity}(List{Entity}Request) returns (List{Entity}Response);
}

// 消息定义
message {Entity}Info {
  uint32 id = 1;
  string name = 2 [(validate.rules).string = {min_len: 1, max_len: 50}];
  int32 status = 3;
  google.protobuf.Timestamp created_at = 4;
}
```

## 2. 命名规范

| 类型 | 命名规则 | 示例 |
|------|----------|------|
| Service | PascalCase | `service User` |
| Message | PascalCase | `message UserInfo` |
| RPC | PascalCase | `rpc CreateUser` |
| Field | snake_case | `string user_name` |
| Enum | SCREAMING_SNAKE | `STATUS_ACTIVE` |

## 3. 标准 CRUD 模板

```protobuf
// 创建请求
message CreateUserRequest {
  string username = 1 [(validate.rules).string = {min_len: 3, max_len: 50}];
  string password = 2 [(validate.rules).string = {min_len: 6}];
  string phone = 3 [(validate.rules).string = {pattern: "^1[3-9]\\d{9}$"}];
}

// 创建响应
message CreateUserResponse {
  uint32 id = 1;
  string username = 2;
  string phone = 3;
  int32 status = 4;
}

// 查询请求
message GetUserRequest {
  uint32 id = 1 [(validate.rules).uint32 = {gt: 0}];
}

// 更新请求
message UpdateUserRequest {
  uint32 id = 1 [(validate.rules).uint32 = {gt: 0}];
  string name = 2;
  string phone = 3;
}

// 删除请求
message DeleteUserRequest {
  uint32 id = 1 [(validate.rules).uint32 = {gt: 0}];
}

message DeleteUserResponse {
  bool success = 1;
}

// 列表查询
message ListUserRequest {
  int32 page = 1 [(validate.rules).int32 = {gt: 0}];
  int32 page_size = 2 [(validate.rules).int32 = {gt: 0, lte: 100}];
  string keyword = 3;
}

message ListUserResponse {
  repeated UserInfo items = 1;
  int32 total = 2;
  int32 page = 3;
  int32 page_size = 4;
}
```

## 4. 生成代码

```bash
# 生成 proto 代码
cd service/{name}
make api

# 输出文件
api/{name}/v1/
├── {name}.pb.go          # message 结构
├── {name}_grpc.pb.go     # gRPC client/server
└── {name}.proto          # 源文件
```

## 5. HTTP API 映射 (可选)

```protobuf
import "google/api/annotations.proto";

service User {
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse) {
    option (google.api.http) = {
      post: "/v1/users"
      body: "*"
    };
  }
  
  rpc GetUser(GetUserRequest) returns (UserInfo) {
    option (google.api.http) = {
      get: "/v1/users/{id}"
    };
  }
}
```

## 下一步

- [配置管理](03-configuration.md) → 配置结构与加载
