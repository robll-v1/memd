package config

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func DefaultDBPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "./memd.db"
	}
	return filepath.Join(home, ".memd", "memd.db")
}

func EnsureParentDir(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0o755)
}

func StableWorkspaceID(dir string) string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		abs = dir
	}
	sum := sha1.Sum([]byte(abs))
	return fmt.Sprintf("ws_%s", hex.EncodeToString(sum[:])[:12])
}

func ResolveWorkspaceID(explicit, dir string) string {
	if strings.TrimSpace(explicit) != "" {
		return explicit
	}
	return StableWorkspaceID(dir)
}
