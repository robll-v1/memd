# Releasing memd

## Option A: Tag-based release (recommended)

This repository includes `.github/workflows/release.yml`.

When you push a tag like `v0.1.0`, GitHub Actions will:

1. build release binaries,
2. create or update the GitHub Release,
3. attach the binaries as assets.

### Steps

```bash
git checkout main
git pull --ff-only
go test ./...
git tag v0.1.0
git push origin main --tags
```

## Option B: Manual draft release in GitHub UI

If you prefer to publish from GitHub's web UI:

1. create a tag first,
2. push the tag,
3. open the repo's Releases page,
4. publish the release for that tag.

The release workflow can still attach built assets when the tag exists.

## Suggested checklist

- Update README if public usage changed
- Run `gofmt -w ./cmd ./internal`
- Run `go test ./...`
- Confirm `go build ./...` succeeds
- Tag with `v0.x.y`
