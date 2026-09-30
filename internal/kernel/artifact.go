package kernel

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// HashBytes returns the lowercase hex SHA-256 of b.
func HashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// Store is a local content-addressed artifact directory.
// Files are named by their content hash. No registry, no interface.
type Store struct {
	Dir string
}

func (s Store) Put(b []byte) (hash string, err error) {
	if err := os.MkdirAll(s.Dir, 0o755); err != nil {
		return "", err
	}
	hash = HashBytes(b)
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

func (s Store) Get(hash string) ([]byte, error) {
	b, err := os.ReadFile(filepath.Join(s.Dir, hash))
	if err != nil {
		return nil, fmt.Errorf("artifact %s: %w", hash, err)
	}
	if got := HashBytes(b); got != hash {
		return nil, fmt.Errorf("artifact %s: bytes hash to %s", hash, got)
	}
	return b, nil
}
