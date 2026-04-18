---
trigger: manual
description: Kratos微服務實戰指南 (六) 啟動與測試 (LearnKu文章64942)
globs: 
---

## Kratos 微服務實戰之用戶服務 (六) 啟動與測試

參考文章: https://learnku.com/articles/64942

---

### 1. 啟動程序

#### 啟動 Consul 服務

```bash
docker run -d -p 8500:8500 -p 8300:8300 -p 8301:8301 -p 8302:8302 -p 8600:8600/udp consul consul agent -dev -client=0.0.0.0
```

瀏覽器訪問 `http://127.0.0.1:8500/ui/dc1/services` 測試是否安裝成功。

---

#### 啟動 Kratos 服務

根目錄 `service/user` 執行命令：

```bash
kratos run
```

或：

```bash
cd cmd/user && go run .
```

> 別忘記根據 data 裡面的 user struct 創建對應的數據庫表，這裡也可以寫一個 gorm 創建表的文件進行創建。

---

### 2. 簡單測試

由於沒寫對外訪問的 http 服務，這裡還沒有加入單元測試，所以先創建個文件鏈接啟動過的 grpc 服務簡單測試一下。

#### 根目錄新建 `user/test/user.go` 文件

```go
package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	v1 "user/api/user/v1"
)

var userClient v1.UserClient
var conn *grpc.ClientConn

func main() {
	Init()
	TestCreateUser() // 創建用戶
	conn.Close()
}

// Init 初始化 grpc 鏈接 注意這裡鏈接的端口
func Init() {
	var err error
	conn, err = grpc.Dial("127.0.0.1:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("grpc link err: " + err.Error())
	}
	userClient = v1.NewUserClient(conn)
}

func TestCreateUser() {
	rsp, err := userClient.CreateUser(context.Background(), &v1.CreateUserInfo{
		Mobile:   fmt.Sprintf("1388888888%d", 1),
		Password: "admin123",
		NickName: fmt.Sprintf("YWWW%d", 1),
	})
	if err != nil {
		panic("grpc 創建用戶失敗: " + err.Error())
	}
	fmt.Println(rsp.Id)
}
```

#### 執行測試

這裡別忘記啟動 kratos user 服務之後，再執行 `test/user.go` 文件：

```bash
# 在 test 目錄下
go run user.go
```

查詢執行結果，是否有一個 ID 輸出，查詢自己的數據庫，看看是否有插入的數據了。

---

### 3. 後續章節

本系列文章還包括：

- [Go-kratos 框架商城微服務實戰之用戶服務 (二) 單元測試](https://learnku.com/articles/64950)
- [Go-kratos 框架商城微服務實戰之用戶服務 (三) 實現 RPC](https://learnku.com/articles/64992)
- [Go-kratos 框架商城微服務實戰之商城服務 (四) BFF 層](https://learnku.com/articles/65095)
- [Go-kratos 框架商城微服務實戰之商城服務 (五) BFF API](https://learnku.com/articles/65241)
- [Go-kratos 框架商城微服務實戰之商品服務 (六) 商品分類](https://learnku.com/articles/65635)
- [Go-kratos 框架商城微服務實戰之商品服務 (七) 商品類型](https://learnku.com/articles/65838)
- [Go-kratos 框架商城微服務實戰之商品服務 (八) 商品規格](https://learnku.com/articles/65865)
- [Go-kratos 框架商城微服務實戰之商品服務 (九) 商品屬性](https://learnku.com/articles/65998)
- [Go-kratos 框架商城微服務實戰之商品服務 (十) 商品創建](https://learnku.com/articles/66175)
- [Go-kratos 框架商城微服務實戰之商品服務 (十一) Elasticsearch 商品搜索](https://learnku.com/articles/66366)
- [Go-kratos 框架商城微服務實戰之購物車服務 (十二)](https://learnku.com/articles/67497)

---

### 參考資源

- **源碼**: https://github.com/aliliin/kratos-shop
- **Kratos 官方文檔**: https://go-kratos.dev/docs/getting-started/start
- **Wire 依賴注入**: https://go-kratos.dev/blog/go-project-wire
- **Project Layout**: https://go-kratos.dev/blog/go-layout-operation-process
- **Go工程化-依賴注入**: https://go-kratos.dev/blog/go-project-wire
