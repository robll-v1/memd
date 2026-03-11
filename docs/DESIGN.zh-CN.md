# memd v1 设计文档

`memd` 是一个面向 AI Agent 的轻量、本地优先记忆运行时。

## 核心特性

- 单一 Go 二进制
- 默认使用本地 SQLite 存储
- REST + MCP 只是同一核心服务层的两种薄适配
- 按 `workspace + agent` 做隔离
- 支持显式纠错和 exact duplicate 清理
- embedding rerank 是可选增强，而不是硬依赖

## v1 范围

- `store`
- `retrieve`
- `search`
- `correct`
- `delete`
- `dedup preview`
- `dedup apply`
- `profile`
- `health`

## v1 不做

- 自动 ingest
- snapshot / branch / reflection
- 图检索
- 完整多租户共享服务
- Web UI

## 默认身份模型

- `workspace_id`：所有数据访问都必须带上
- `agent_id`：写入必须带，读取可选
- `session_id`：仅作为排序和溯源信息，不作为主隔离键

## 默认存储结构

- `memories`
- `memory_events`
- `memories_fts`（SQLite FTS5）

## 检索行为

- FTS5 负责主候选召回
- recency 和 same-agent affinity 参与排序
- 如果配置了 embedding，则在关键词候选集上做进程内 rerank

## 去重行为

- exact duplicate = 同 `workspace_id + kind + content_norm`
- near duplicate = 同一主题，且词法/语义相似度较高
- v1 只自动 apply exact duplicate
