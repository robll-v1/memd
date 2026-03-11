# Installation

For most users, **Go is not required**.

The recommended path is:

1. download a release binary,
2. make it executable,
3. point Codex MCP at that binary.

## 1. Choose the correct binary

From the Releases page:

- `memd-linux-amd64`
- `memd-darwin-amd64`
- `memd-darwin-arm64`
- `memd-windows-amd64.exe`

Choose based on your platform:

- macOS Intel → `memd-darwin-amd64`
- macOS Apple Silicon → `memd-darwin-arm64`
- Linux x86_64 → `memd-linux-amd64`
- Windows x86_64 → `memd-windows-amd64.exe`

## 2. Install locally

### macOS / Linux

```bash
chmod +x ./memd-darwin-amd64
mkdir -p ~/bin
mv ./memd-darwin-amd64 ~/bin/memd
~/bin/memd doctor
```

Adjust the filename if you downloaded a different platform build.

### Windows

Rename the binary if needed and place it in a stable path, for example:

- `C:\Tools\memd\memd.exe`

Then run:

```powershell
C:\Tools\memd\memd.exe doctor
```

## 3. Optional: build from source

Only needed if you want to build `memd` yourself.

Requirements:

- Go 1.24+

```bash
git clone https://github.com/robll-v1/memd.git
cd memd
go build -o memd ./cmd/memd
```

## 4. Verify the binary

Run:

```bash
memd doctor
```

Expected result:

- a JSON object
- `embedding_enabled: false` unless you configured embeddings
- `total: 0` on a fresh database

## 5. Next step

Continue with:

- `docs/CODEX.md` for Codex MCP setup
