# 贡献指南

## 开发

```bash
go test ./...
go run ./cmd/memd doctor
go run ./cmd/memd serve
```

## 项目原则

- 默认保持 `memd` 本地优先、轻量可运行。
- 不要让 embedding 成为正确性的硬依赖。
- 优先采用显式纠错和 exact dedup，不要做过度激进的自动合并。
- 保持 REST 和 MCP 只是核心服务层的薄适配层。

## 发版前检查

- 运行 `gofmt -w ./cmd ./internal`
- 运行 `go test ./...`
- 确认 `go build ./...` 成功
- 如果对外接口或 CLI 参数有变化，更新文档
