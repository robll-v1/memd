# 发布说明

## 构建

```bash
go build -o memd ./cmd/memd
```

## 本地冒烟检查

```bash
go test ./...
./memd doctor
./memd serve
./memd mcp --agent codex
```

## Codex MCP 配置示例

```toml
[mcp_servers.memd]
command = "/absolute/path/to/memd"
args = ["mcp", "--agent", "codex"]
```

## GitHub 发布流程

1. 提交所有改动。
2. 创建或连接一个公开的 GitHub 仓库。
3. 推送 `main`。
4. 使用 `v0.x.y` 形式打 tag。
5. 之后可以继续补二进制分发、Homebrew 或其他发布方式。
