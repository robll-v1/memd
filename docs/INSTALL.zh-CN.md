# 安装说明

对大多数用户来说，**不需要安装 Go**。

推荐路径是：

1. 直接下载 release 二进制，
2. 赋予执行权限，
3. 在 Codex MCP 配置里指向这个二进制。

## 1. 选择正确的二进制

在 Releases 页面里你会看到：

- `memd-linux-amd64`
- `memd-darwin-amd64`
- `memd-darwin-arm64`
- `memd-windows-amd64.exe`

按平台选择：

- macOS Intel → `memd-darwin-amd64`
- macOS Apple Silicon → `memd-darwin-arm64`
- Linux x86_64 → `memd-linux-amd64`
- Windows x86_64 → `memd-windows-amd64.exe`

## 2. 本地安装

### macOS / Linux

```bash
chmod +x ./memd-darwin-amd64
mkdir -p ~/bin
mv ./memd-darwin-amd64 ~/bin/memd
~/bin/memd doctor
```

如果你下载的是别的平台文件名，请替换为对应文件名。

### Windows

把二进制放到一个固定目录，例如：

- `C:\Tools\memd\memd.exe`

然后运行：

```powershell
C:\Tools\memd\memd.exe doctor
```

## 3. 可选：从源码构建

只有你想自己编译时才需要这条路径。

要求：

- Go 1.24+

```bash
git clone https://github.com/robll-v1/memd.git
cd memd
go build -o memd ./cmd/memd
```

## 4. 验证二进制

运行：

```bash
memd doctor
```

预期结果：

- 输出一个 JSON 对象
- 如果没有配置 embedding，`embedding_enabled` 应为 `false`
- 对全新数据库，`total` 应为 `0`

## 5. 下一步

继续看：

- `docs/CODEX.zh-CN.md`
