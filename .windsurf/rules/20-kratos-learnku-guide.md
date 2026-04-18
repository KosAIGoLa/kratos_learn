---
trigger: manual
description: Kratos微服務實戰指南 (一) 項目初始化與API定義 (LearnKu文章64942)
globs: 
---

## Kratos 微服務實戰之用戶服務 (一) 初識 Kratos 框架

參考文章: https://learnku.com/articles/64942

> 注意：現在 kratos 框架有新版本發布，部分代碼有出入，一切以官方文檔為準，不過基本結構是差不多的。
> 推薦先看 [Kratos 官方文檔](https://go-kratos.dev/docs/getting-started/start) 更加流暢觀看此文章。
>
> 本機需要已安裝好 kratos、proto、wire、make 等所需的命令工具。

---

### 1. 初始化項目目錄

進入自己電腦中存放 Go 項目的目錄，新建目錄並進入：

```bash
mkdir -p kratos-shop/service
cd kratos-shop/service
```

執行 kratos new user 命令並進入 user 目錄：

```bash
kratos new user
cd user
```

執行命令添加 proto 文件：

```bash
kratos proto add api/user/v1/user.proto
```

這時你在 `kratos-shop/service/user/api/user/v1` 目錄下會看到新的 `user.proto` 文件已經創建好了。

接下來執行命令生成對應的 service 文件：

```bash
kratos proto server api/user/v1/user.proto -t internal/service
```

刪除不需要的 proto 文件和 service 文件：

```bash
rm -rf api/helloworld/
rm internal/service/greeter.go
```

#### 完整的命令代碼如下

```bash
mkdir -p kratos-shop/service && cd kratos-shop/service
kratos new user
cd user
kratos proto add api/user/v1/user.proto
kratos proto server api/user/v1/user.proto -t internal/service
rm -rf api/helloworld/
rm internal/service/greeter.go
```

---

### 2. 定義 API 接口 (Proto)

#### 修改 `api/user/v1/user.proto` 文件

```protobuf
syntax = "proto3";

package user.v1;

option go_package = "user/api/user/v1;v1";

service User {
  rpc CreateUser(CreateUserInfo) returns (UserInfoResponse);
}

// 創建用戶所需字段
message CreateUserInfo {
  string nickName = 1;
  string password = 2;
  string mobile   = 3;
}

// 返回用戶信息
message UserInfoResponse {
  int64  id       = 1;
  string password = 2;
  string mobile   = 3;
  string nickName = 4;
  int64  birthday = 5;
  string gender   = 6;
  int32  role     = 7;
}
```

> proto 基本的語法請自行學習，目前這裡只先提供了一個創建用戶的 rpc 接口，後續會逐步添加其他 rpc 接口。

#### 生成 user.proto 定義的接口信息

進入到 `service/user` 目錄下，執行 make api 命令：

```bash
cd user
make api
```

這時可以看到 `user/api/user/v1/` 目錄下多出了 proto 創建的文件：

```
├── api
│   └── user
│       └── v1
│           ├── user.pb.go
│           ├── user.proto
│           └── user_grpc.pb.go
```

---

### 下一步

完成項目初始化和 API 定義後，請參考：

- `21-kratos-learnku-config.md` - 配置管理
- `22-kratos-learnku-data.md` - 數據層實現
- `23-kratos-learnku-biz-service.md` - 業務層與服務層
- `24-kratos-learnku-server.md` - 服務器配置與啟動
- `25-kratos-learnku-testing.md` - 測試

---

### 參考資源

- 源碼: https://github.com/aliliin/kratos-shop
- Kratos 官方文檔: https://go-kratos.dev/docs/getting-started/start
- Wire 依賴注入: https://go-kratos.dev/blog/go-project-wire
- Project Layout: https://go-kratos.dev/blog/go-layout-operation-process