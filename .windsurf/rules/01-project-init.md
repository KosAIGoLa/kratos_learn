---
trigger: manual
description: Kratos 项目初始化与创建
globs: ["**/cmd/**/main.go"]
---

# Kratos 项目初始化

## 1. 创建新项目

```bash
# 进入服务目录
mkdir -p service && cd service

# 创建服务
kratos new {service-name}
cd {service-name}
```

## 2. 定义 API 接口

```bash
# 添加 proto 文件
kratos proto add api/{name}/v1/{name}.proto

# 生成 service 骨架
kratos proto server api/{name}/v1/{name}.proto -t internal/service

# 清理示例文件
rm -rf api/helloworld/
rm -f internal/service/greeter.go
rm -f internal/biz/greeter.go
rm -f internal/data/greeter.go
```

## 3. 完整初始化命令

```bash
mkdir -p service && cd service
kratos new {service-name}
cd {service-name}
kratos proto add api/{name}/v1/{name}.proto
kratos proto server api/{name}/v1/{name}.proto -t internal/service

# 清理示例
rm -rf api/helloworld/
rm -f internal/service/greeter.go
rm -f internal/biz/greeter.go
rm -f internal/data/greeter.go
```

## 4. 修改 Makefile

添加 wire 命令：

```makefile
wire:
	cd cmd/{name}/ && wire
```

## 5. 项目结构检查

初始化完成后应有以下结构：

```
service/{name}/
├── api/{name}/v1/{name}.proto     ✅ 新增
├── cmd/{name}/
│   ├── main.go
│   ├── wire.go
│   └── wire_gen.go
├── configs/config.yaml
├── internal/
│   ├── biz/
│   ├── conf/
│   ├── data/
│   ├── server/
│   └── service/
└── go.mod
```

## 下一步

- [API 定义规范](02-api-definition.md) → 定义 protobuf 接口
