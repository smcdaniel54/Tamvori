package kernel

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
)

// HashBytes returns the lowercase hex SHA-256 of b.
func HashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// store is a local content-addressed artifact directory.
type store struct {
	Dir string
}

func (s store) put(b []byte) (string, error) {
	if err := os.MkdirAll(s.Dir, 0o755); err != nil {
		return "", err
	}
	hash := HashBytes(b)
	path := filepath.Join(s.Dir, hash)
	if _, err := os.Stat(path); err == nil {
		return hash, nil
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return "", err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return "", err
	}
	return hash, nil
}
