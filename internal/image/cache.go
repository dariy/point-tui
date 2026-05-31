package image

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type cacheKey struct {
	path string
	cols int
	rows int
}

// Cache stores rendered ANSI strings keyed by (path, cols, rows) in memory,
// and raw image bytes keyed by path on disk.
type Cache struct {
	mu      sync.Mutex
	ansi    map[cacheKey]string
	diskDir string
}

// NewCache creates a Cache that stores raw bytes under diskDir.
func NewCache(diskDir string) *Cache {
	return &Cache{
		ansi:    make(map[cacheKey]string),
		diskDir: diskDir,
	}
}

// GetANSI returns a cached ANSI render, or ("", false) if not cached.
func (c *Cache) GetANSI(path string, cols, rows int) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.ansi[cacheKey{path, cols, rows}]
	return v, ok
}

// PutANSI caches an ANSI render.
func (c *Cache) PutANSI(path string, cols, rows int, ansi string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ansi[cacheKey{path, cols, rows}] = ansi
}

// InvalidateANSI clears all ANSI renders for a given path (e.g. on resize).
func (c *Cache) InvalidateANSI(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k := range c.ansi {
		if k.path == path {
			delete(c.ansi, k)
		}
	}
}

func diskName(path string) string {
	h := sha256.Sum256([]byte(path))
	return fmt.Sprintf("%x", h)
}

// GetRaw returns raw image bytes from disk cache, or (nil, false).
func (c *Cache) GetRaw(path string) ([]byte, bool) {
	if c.diskDir == "" {
		return nil, false
	}
	p := filepath.Join(c.diskDir, diskName(path))
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, false
	}
	return data, true
}

// PutRaw writes raw image bytes to disk cache.
func (c *Cache) PutRaw(path string, data []byte) error {
	if c.diskDir == "" {
		return nil
	}
	if err := os.MkdirAll(c.diskDir, 0o700); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(c.diskDir, diskName(path)), data, 0o600)
}
