# util — 项目工具

本目录存放与本仓库相关的辅助工具。

## scaffold — 项目初始化脚手架

按照 `kbfs_prediction` 的目录约定（`biz`(handler/app/domain) + `cmd/services` + `infra` + `server` + `pkg` + `configs` + `docker`），一键生成一个新的 Kratos 微服务骨架。

生成的骨架已集成拳游内部 `core-sdk-go` 基础设施：

- **Nacos 配置中心 + 服务注册**（`infra/nacos`）
- **Redis 客户端 + asynq 异步任务队列**（`infra/redis`、`infra/task`，含 producer/consumer 示例）
- **RabbitMQ 消息队列**（`infra/rabbitmq`，含 producer/consumer 示例）
- **zaplog 日志**（`infra/logger`）
- **配置加载**（`infra/conf`，含 Nacos / Log / RuntimeMonitor）
- **日志**（`infra/logger`，zaplog）
- **健康检查 + Prometheus 指标**（`infra/health`、`infra/metrics`）
- **MySQL + gorm_gen 数据库模型生成**（`infra/mysql`、`cmd/custom`）
- **MongoDB 数据访问**（`infra/mongoDB`）
- **RPC 客户端（Nacos 服务发现）**（`infra/rpc`，含 user / sports_meta 示例）
- **定时任务（XXL-Job）**（`biz/cron`，仅在 task 服务中装配）
- 同时生成 **api（生产者侧）** 与 **task（消费者侧）** 两个服务入口。

### 使用

```bash
# 在本仓库根目录执行
go run ./ \
  -module gl.quanyougame.net/backend/kbfs_demo \
  -name kbfs_demo \
  -out ../kbfs_demo
```

参数说明：

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `-module` | 是 | 新项目的 Go module 路径，如 `gl.quanyougame.net/backend/kbfs_demo` |
| `-name`   | 否 | 服务短名，默认取 module 最后一段，用于二进制名 / 容器名 |
| `-server` | 否 | Kratos 服务注册名，默认 `qy.kbfs.<去掉 kbfs_ 前缀的 name>` |
| `-out`    | 否 | 生成目录，默认 `./<name>` |
| `-force`  | 否 | 目标目录已存在同名文件时是否覆盖，默认 `false` |

### 生成后的目录结构

```bash
<out>/
├── cmd/
│   ├── custom/            # gorm_gen 代码生成入口
│   └── services/
│       ├── api/           # API 服务（生产者侧）：main + wire + wire_gen
│       └── task/          # Task 服务（消费者侧）：main + wire + wire_gen
├── biz/
│   ├── handler/           # 接入层（HTTP / gRPC）
│   ├── app/               # 应用层（用例编排）
│   ├── domain/            # 领域层（核心业务）
│   ├── cron/              # 定时任务（XXL-Job）
│   └── consts/            # 常量
├── infra/
│   ├── conf/              # 配置加载与定义（Nacos / Log / RuntimeMonitor）
│   ├── logger/            # zaplog 日志
│   ├── health/            # 健康检查 /ping
│   ├── metrics/           # Prometheus + OTel 指标
│   ├── nacos/             # 服务注册 + 配置中心
│   ├── redis/             # redis 集群客户端
│   ├── task/              # asynq 任务（client/server + producer/consumer）
│   ├── rabbitmq/          # rabbitmq（client + producer/consumer）
│   ├── mysql/             # MySQL 连接 + dao + po/query（gorm_gen 输出）
│   ├── mongoDB/           # MongoDB 连接 + dao
│   └── rpc/               # 外部 RPC 客户端（Nacos 服务发现）
├── pkg/utils/             # 通用工具
├── server/                # HTTP / gRPC server 装配
├── configs/               # 分环境配置 config_<env>.yaml
├── docker/                # docker-compose + init.sql
├── Dockerfile
├── Makefile
├── go.mod
├── README.md
└── .gitignore
```

### 生成后的下一步

```bash
cd <out>
# 私有模块需要配置 GOPRIVATE（拳游内部 git）
export GOPRIVATE=glprivate.quanyougame.net,gl.quanyougame.net
go mod tidy
make generate   # 用 wire 重新生成依赖注入代码（可选，已附带可用的 wire_gen.go）
make gorm_gen   # 从数据库表生成 gorm model/query（需先配置 cmd/custom/custom.go）
make run-api    # 启动 API 服务（生产者侧）
make run-task   # 启动 Task 服务（消费者侧）
```

启动后可访问：

- `GET http://localhost:8000/ping` → `success`（健康检查）
- `GET http://localhost:8000/metrics` → Prometheus 指标
- `GET http://localhost:8000/hello/cursor` → `{"message":"hello, cursor"}`

### 设计说明

- 脚手架本身是单文件 Go 程序（`scaffold/main.go`），模板通过 `go:embed` 内嵌于 `scaffold/templates/`，无需额外依赖即可运行。
- 模板使用 `[[ ]]` 作为分隔符，避免与配置文件中的 `${...}` 及 Go 代码冲突。
- 生成的骨架接入了拳游内部 `core-sdk-go`（nacos / redis / task / rabbitmq / zaplog），与 `kbfs_prediction` 的 `infra/` 实现保持一致。**因此 `go mod tidy` / 编译需要能访问内部私有模块仓库（配置 `GOPRIVATE`）**，公司内网环境开箱即用。
- api 与 task 复用同一套 `biz` / `infra`，仅 `cmd/services/*/wire.go` 装配不同 provider：api 侧装配生产者（rabbitmq producer + asynq task client），task 侧额外装配消费者（rabbitmq consumer + asynq task server）。
