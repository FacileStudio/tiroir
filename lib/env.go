// Package env implements an encrypted key-value store of shell variables.
//
// Values are held at rest as an AES-256-GCM ciphertext blob keyed by a separate
// random key file. The split keeps the store greppable-safe: a naive scanner
// hunting for "KEY=value" strings in dotenv files finds only opaque bytes. This
// is an anti-lazy-scanner layer, not a real security boundary: anything that
// can run the tiroir binary or read the key file reads the values anyway.
package env

import (
	"os"
	"path/filepath"
	"sort"
)

// EnvStore reads and writes the encrypted variable store for an invoking user.
type EnvStore struct {
	storePath string
	keyPath   string
}

// New returns an EnvStore rooted at the invoking user's home directory.
func New() (*EnvStore, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return NewAt(home), nil
}

// NewAt returns an EnvStore whose files live in dir.
func NewAt(dir string) *EnvStore {
	return &EnvStore{
		storePath: filepath.Join(dir, storeName),
		keyPath:   filepath.Join(dir, keyName),
	}
}

// Get returns the value for key and whether it was present.
func (s *EnvStore) Get(key string) (string, bool, error) {
	vals, err := s.read()
	if err != nil {
		return "", false, err
	}
	v, ok := vals[key]
	return v, ok, nil
}

// List returns all keys currently in the store.
func (s *EnvStore) List() ([]string, error) {
	vals, err := s.read()
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(vals))
	for k := range vals {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, nil
}

// Export returns the store as "KEY=value" lines, one per key, sorted by key.
func (s *EnvStore) Export() ([]string, error) {
	vals, err := s.read()
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(vals))
	for k := range vals {
		out = append(out, k+"="+vals[k])
	}
	sort.Strings(out)
	return out, nil
}

// Set writes key to value, preserving all other keys.
func (s *EnvStore) Set(key, value string) error {
	_, err := s.mutate(func(vals map[string]string) map[string]string {
		vals[key] = value
		return vals
	})
	return err
}

// Delete removes key from the store. Removing a missing key is not an error.
func (s *EnvStore) Delete(key string) error {
	_, err := s.mutate(func(vals map[string]string) map[string]string {
		delete(vals, key)
		return vals
	})
	return err
}

func (s *EnvStore) mutate(f func(map[string]string) map[string]string) (map[string]string, error) {
	vals, err := s.read()
	if err != nil {
		return nil, err
	}
	vals = f(vals)
	return vals, s.save(vals)
}