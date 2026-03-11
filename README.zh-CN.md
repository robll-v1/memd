# memd

[![CI](https://github.com/robll-v1/memd/actions/workflows/ci.yml/badge.svg)](https://github.com/robll-v1/memd/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/robll-v1/memd)](./LICENSE)

面向 AI Agent 的本地优先记忆运行时。

English README: `README.md`

`memd` 是一个轻量记忆层，面向编码类 agent。它把长期记忆保存在本地 SQLite 中，提供本地 REST API，并可通过 MCP 挂载到 agent 客户端。

## 特性

- 默认使用 SQLite（`~/.memd/memd.db`）
- 单二进制本地运行
- REST API
- 面向 Codex 等客户端的 stdio MCP server
- 显式纠错链路
- exact duplicate 预览 / 清理
- 可选的 OpenAI 兼容 embedding rerank
- 核心功能不依赖 LLM

## 安装

### 方式 A：直接下载 release 二进制

最推荐的方式是不装 Go，直接下载 release 二进制：

- Releases：`https://github.com/robll-v1/memd/releases`

下载后赋予执行权限：

```bash
chmod +x ./memd-darwin-amd64
mv ./memd-darwin-amd64 ./memd
```

### 方式 B：从源码构建

要求：

- Go 1.24+

```bash
git clone https://github.com/robll-v1/memd.git
cd memd
go build -o memd ./cmd/memd
```

## 快速开始

### 本地健康检查

```bash
./memd doctor
```

### 启动本地 REST 服务

```bash
./memd serve
```

### 通过 REST 写入并检索一条记忆

```bash
curl -X POST http://127.0.0.1:8081/v1/memories \
  -H 'Content-Type: application/json' \
  -d '{
    "workspace_id": "demo",
    "agent_id": "codex",
    "kind": "fact",
    "content": "MatrixOne runs locally on port 6001"
  }'

curl -X POST http://127.0.0.1:8081/v1/memories/retrieve \
  -H 'Content-Type: application/json' \
  -d '{
    "workspace_id": "demo",
    "agent_id": "codex",
    "query": "what port does MatrixOne use",
    "limit": 5
  }'
```

## Codex 配置

`Codex CLI` 和 `Codex for VS Code` 共用同一份 MCP 配置文件：

- `~/.codex/config.toml`

你只需要配置一次 `memd`。

详细说明见：`docs/CODEX.zh-CN.md`

推荐配置如下：

```toml
[mcp_servers.memd]
command = "/absolute/path/to/memd"
args = ["mcp", "--agent", "codex"]
```

示例：

```toml
[mcp_servers.memd]
command = "/Users/you/bin/memd"
args = ["mcp", "--agent", "codex"]
```

### 重要说明

- 不要先手工启动一个长期前台运行的 `memd mcp`
- 应该让 Codex 通过 `~/.codex/config.toml` 自己拉起 MCP server
- 改完配置后，要开一个**新的** Codex 会话
- 已经打开的 Codex 会话不会热加载新的 MCP 配置

## 命令

```bash
go run ./cmd/memd serve
go run ./cmd/memd mcp --agent codex
go run ./cmd/memd doctor
```

## REST 接口

- `POST /v1/memories`
- `POST /v1/memories/retrieve`
- `GET /v1/memories/search`
- `PUT /v1/memories/{id}/correct`
- `DELETE /v1/memories/{id}`
- `POST /v1/memories/dedup/preview`
- `POST /v1/memories/dedup/apply`
- `GET /v1/profile`
- `GET /v1/health`

## 说明

- REST 调用必须显式传 `workspace_id`
- MCP 模式下，默认 `workspace_id` 来自当前工作目录哈希
- MCP 模式下，`agent_id` 默认来自 `--agent` 参数
- v1 只会自动清理 exact duplicates
- near-duplicate 在 v1 里只预览，不自动删除
- embedding 是可选增强项，不影响核心功能可用性

## 排错

### Codex 看不到 `memd`

- 检查 `~/.codex/config.toml`
- 确认 `command` 指向真实存在的绝对路径
- 修改配置后，重新开一个新的 Codex 会话

### `memory_health` 能用，但检索为空

- 先存一条记忆再测
- 如果依赖默认 MCP `workspace_id`，确认你在同一个项目目录里
- 如果走 REST，确认 `workspace_id` 和写入时保持一致

### 去重没有删除 near-duplicate

这是 v1 设计如此：

- `memory_dedup_apply` 只自动处理 exact duplicate
- near-duplicate 只做 preview，不自动删除

## 仓库文档

- 设计文档（英文）：`docs/DESIGN.md`
- 设计文档（中文）：`docs/DESIGN.zh-CN.md`
- 安装说明（英文）：`docs/INSTALL.md`
- 安装说明（中文）：`docs/INSTALL.zh-CN.md`
- Codex 配置（英文）：`docs/CODEX.md`
- Codex 配置（中文）：`docs/CODEX.zh-CN.md`
- 发布说明（英文）：`docs/PUBLISHING.md`
- 发布说明（中文）：`docs/PUBLISHING.zh-CN.md`
- 发版流程（英文）：`docs/RELEASING.md`
- 发版流程（中文）：`docs/RELEASING.zh-CN.md`
- 贡献指南（英文）：`CONTRIBUTING.md`
- 贡献指南（中文）：`CONTRIBUTING.zh-CN.md`
