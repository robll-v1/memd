# Codex 配置说明

本文档说明如何在以下客户端中使用 `memd`：

- Codex CLI
- Codex for VS Code

这两个客户端共用同一份 MCP 配置文件：

- `~/.codex/config.toml`

## 1. 安装或准备 `memd`

如果你是从源码构建：

```bash
go build -o memd ./cmd/memd
```

如果你不想装 Go，直接去 Releases 下载对应平台的二进制即可：

- `https://github.com/robll-v1/memd/releases`

## 2. 添加 MCP 配置

编辑 `~/.codex/config.toml`，加入：

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

## 3. 新开一个 Codex 会话

注意：

- 已经打开的会话不会热加载新的 MCP server
- 修改完配置以后，要开一个全新的 Codex 会话

## 4. 冒烟测试

在新的 Codex 会话里，依次执行这些提示词：

```text
Please call memory_health first.
```

```text
Please store this fact with memd: my MatrixOne port is 6001.
```

```text
Please retrieve what port my MatrixOne uses.
```

## 5. 精确重复测试

把同一条事实存两次，然后让 Codex 执行：

```text
Please run memory_dedup_preview and show exact duplicate groups.
```

接着执行：

```text
Please run memory_dedup_apply on the exact duplicate group.
```

## 6. 可选 embedding 配置

如果你想启用语义 rerank，可以在 MCP 命令里带上 embedding 参数。

例如：

```toml
[mcp_servers.memd]
command = "/absolute/path/to/memd"
args = [
  "mcp",
  "--agent", "codex",
  "--embed-provider", "openai",
  "--embed-api-key", "sk-...",
  "--embed-model", "text-embedding-3-small"
]
```

## 说明

- MCP 模式下，默认 `workspace_id` 会根据当前工作目录生成
- REST 模式下，必须显式传 `workspace_id`
- v1 中 `memory_dedup_apply` 只会自动处理 exact duplicate
