# memd 发版流程

## 方式 A：基于 tag 的自动发布（推荐）

仓库已经包含 `.github/workflows/release.yml`。

当你推送一个类似 `v0.1.0` 的 tag 时，GitHub Actions 会自动：

1. 构建发布二进制，
2. 创建或更新 GitHub Release，
3. 把二进制挂到 Release Assets 中。

### 操作步骤

```bash
git checkout main
git pull --ff-only
go test ./...
git tag v0.1.0
git push origin main --tags
```

## 方式 B：在 GitHub UI 中手工发版

如果你更喜欢在 GitHub 网页上完成发布：

1. 先创建 tag，
2. 推送 tag，
3. 打开仓库的 Releases 页面，
4. 为该 tag 发布 release。

即便你走手工发布路径，release workflow 依然可以在 tag 存在时自动附加构建产物。

## 建议检查项

- 如果对外使用方式变了，更新 README
- 运行 `gofmt -w ./cmd ./internal`
- 运行 `go test ./...`
- 确认 `go build ./...` 成功
- 使用 `v0.x.y` 形式打 tag
