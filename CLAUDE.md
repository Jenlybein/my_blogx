# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

这是一个基于 Go + Gin + GORM 的博客后端服务系统，支持文章管理、用户认证、全文搜索、云存储等功能。

## 常用命令

### 运行服务
```bash
# 使用默认配置文件 (settings.yaml)
go run main.go

# 指定配置文件
go run main.go -f custom_settings.yaml
```

### 数据库迁移
```bash
# 运行数据库迁移（创建/更新表结构）
go run main.go -db

# 指定配置文件并迁移
go run main.go -f settings.yaml -db
```

### Elasticsearch 索引初始化
```bash
# 初始化 ES 索引
go run main.go -es
```

### 用户管理
```bash
# 创建用户
go run main.go -t user -s create
```

### 构建项目
```bash
go build -o blogx_server main.go
```

## 项目架构

### 分层架构
项目采用经典的三层架构设计：

- **API 层** (`api/`) - 处理 HTTP 请求，绑定参数，调用 Service 层
- **Service 层** (`service/`) - 业务逻辑处理，数据库操作，缓存管理
- **Model 层** (`models/`) - 数据模型定义（GORM 模型）

### 目录结构说明

```
blogx_server/
├── api/                 # API 处理层
│   ├── article_api/     # 文章相关 API
│   ├── user_api/        # 用户相关 API
│   ├── image_api/       # 图片上传 API
│   └── ...
├── service/             # 业务逻辑层
│   ├── redis_service/   # Redis 缓存服务
│   │   ├── redis_article/  # 文章相关缓存
│   │   └── redis_jwt/      # JWT 缓存
│   ├── es_service/      # Elasticsearch 服务
│   ├── river_service/   # MySQL -> ES 实时同步服务
│   └── ...
├── models/              # 数据模型层 (GORM)
│   ├── article_model.go
│   ├── user_model.go
│   └── ...
├── router/              # 路由注册
├── middleware/           # 中间件
│   ├── auth_middleware.go      # JWT 认证
│   ├── log_middleware.go       # 日志记录
│   └── ...
├── core/                # 核心初始化
│   ├── init_db.go       # 数据库初始化（含读写分离和缓存）
│   ├── init_redis.go    # Redis 初始化
│   ├── init_es.go       # Elasticsearch 初始化
│   └── init_logrus.go    # 日志初始化
├── conf/                # 配置结构体定义
├── flags/               # 命令行参数处理
├── global/              # 全局变量 (DB, Redis, Logger, ESClient 等)
├── store/               # 存储层
│   └── email_store/     # 邮箱验证码存储
├── init/                # 初始化脚本
│   ├── ipbase/          # IP 地址库初始化
│   └── deploy/          # 部署脚本
└── utils/               # 工具函数
```

## 核心功能模块

### 1. 数据库层
- **读写分离**：使用 `gorm.io/plugin/dbresolver` 实现主从分离
  - 配置文件中第一个数据库为主库（写），其余为从库（读）
  - 见 [core/init_db.go:130-146](core/init_db.go#L130-L146)
- **GORM 缓存插件**：使用 `github.com/go-gorm/caches/v4` + Redis 实现查询缓存
  - 自定义 `redisCacher` 实现缓存接口
  - 默认缓存时间：300 秒
  - 见 [core/init_db.go:21-159](core/init_db.go#L21-L159)

### 2. Elasticsearch 集成
- **全文搜索**：基于 Elasticsearch 7.x
- **River 同步服务**：使用 `go-mysql` 实现 MySQL 到 ES 的实时同步
  - 监听 MySQL binlog，自动同步数据变更
  - 配置见 `settings.yaml` 的 `river` 部分
  - 见 [service/river_service/river.go](service/river_service/river.go)

### 3. Redis 缓存策略
- **Service 层缓存**：按模块划分缓存逻辑
  - `redis_article/` - 文章相关缓存（浏览量、点赞数等）
  - `redis_jwt/` - JWT 黑名单缓存
- **GORM 二级缓存**：数据库查询结果自动缓存

### 4. 认证与授权
- **JWT 认证**：使用 `github.com/golang-jwt/jwt`
- **中间件**：`middleware/auth_middleware.go` 处理 JWT 验证
- **登录方式**：
  - 用户名密码登录
  - 邮箱验证码登录
  - QQ 登录（可选）

### 5. 文件存储
- **本地存储**：`uploads/` 目录
- **七牛云存储**：可选的云存储方案
- 配置见 `settings.yaml` 的 `qiniu` 部分

### 6. IP 地址库
- 使用 `ip2region` 实现精确到城市的 IP 定位
- 初始化脚本：`init/ipbase/`

## 配置说明

主配置文件：`settings.yaml`

关键配置项：
- `system` - 服务端口、运行模式
- `db` - 数据库连接（支持多个，实现读写分离）
- `gorm` - GORM 配置（连接池、调试模式）
- `redis` - Redis 连接配置
- `es` - Elasticsearch 连接配置
- `river` - MySQL -> ES
- `jwt` - JWT 密钥和过期时间
- `upload` - 文件上传配置
- `qiniu` - 七牛云配置

## 开发规范

### 新增 API 流程
1. 在 `api/` 对应模块下创建 API 文件
2. 在 `models/` 下创建/使用对应的 Model
3. 在 `service/` 下创建 Service 层处理业务逻辑
4. 在 `router/` 下注册路由
5. 根据需要使用中间件（认证、日志等）

### 缓存开发规范
- 优先使用 Service 层的 Redis 缓存（`service/redis_service/`）
- 对于高频查询，可利用 GORM 缓存插件自动缓存
- 缓存失效策略：数据变更时主动清除缓存

### 数据库操作规范
- 使用 GORM 的 `global.DB` 进行操作
- 默认使用读写分离，写操作自动路由到主库
- 复杂查询考虑使用缓存优化

## 技术栈

- **Web 框架**：Gin v1.11.0
- **ORM**：GORM v1.31.1
- **数据库**：MySQL（支持读写分离）
- **缓存**：Redis v8.11.5
- **搜索引擎**：Elasticsearch v7
- **日志**：Logrus v1.9.3
- **JWT**：golang-jwt/jwt v3
- **数据同步**：go-mysql（MySQL binlog 监听）
- **配置解析**：YAML v3
