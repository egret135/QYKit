# util — 项目工具

本目录存放与本仓库相关的辅助工具。

## scaffold — 项目初始化脚手架

按照 `kbfs_prediction` 的目录约定（`biz`(handler/app/domain) + `cmd/services` + `infra` + `server` + `pkg` + `configs` + `docker`），一键生成一个新的 Kratos 微服务骨架。

### 使用

```bash
# 在本仓库根目录执行
go run ./util/scaffold \
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
├── cmd/services/api/      # 服务入口：main.go + wire.go + wire_gen.go
├── biz/
│   ├── handler/           # 接入层（HTTP / gRPC）
│   ├── app/               # 应用层（用例编排）
│   ├── domain/            # 领域层（核心业务）
│   └── consts/            # 常量
├── infra/conf/            # 配置加载与定义
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
go mod tidy
make generate   # 用 wire 重新生成依赖注入代码（可选，已附带可用的 wire_gen.go）
make run-api    # 本地启动，默认读取 configs/config_dev.yaml
```

启动后可访问：

- `GET http://localhost:8000/ping` → `{"status":"ok"}`
- `GET http://localhost:8000/hello/cursor` → `{"message":"hello, cursor"}`

### 设计说明

- 脚手架本身是单文件 Go 程序（`scaffold/main.go`），模板通过 `go:embed` 内嵌于 `scaffold/templates/`，无需额外依赖即可运行。
- 模板使用 `[[ ]]` 作为分隔符，避免与配置文件中的 `${...}` 及 Go 代码冲突。
- 生成的骨架默认使用**公共依赖**（kratos / wire），保证 `go mod tidy` 后可直接编译运行。接入组织内部的 Nacos 配置中心、MySQL/Redis/RabbitMQ 等私有 SDK 时，参考 `kbfs_prediction` 的 `infra/` 实现替换对应模块即可。
