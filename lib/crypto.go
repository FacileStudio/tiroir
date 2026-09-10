package env

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	filePerm  = 0o600
	storeName = ".tiroir"
	keyName   = ".tiroir.key"
)

var (
	magic     = []byte("TIROIR")
	version   = byte(1)
	keyID     = byte(1)
	headerLen = len(magic) + 2
)

func (s *EnvStore) read() (map[string]string, error) {
	key, err := s.loadOrCreateKey()
	if err != nil {
		return nil, err
	}
	blob, err := os.ReadFile(s.storePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	return decrypt(blob, key)
}

func (s *EnvStore) save(vals map[string]string) error {
	key, err := s.loadOrCreateKey()
	if err != nil {
		return err
	}
	plain, err := json.Marshal(vals)
	if err != nil {
		return err
	}
	blob, err := seal(plain, key)
	if err != nil {
		return err
	}
	return atomicWrite(s.storePath, blob)
}

func (s *EnvStore) loadOrCreateKey() ([]byte, error) {
	if key, err := os.ReadFile(s.keyPath); err == nil {
		return key, nil
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	if err := atomicWrite(s.keyPath, key); err != nil {
		return nil, err
	}
	return key, nil
}

func seal(plain, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	out := make([]byte, 0, headerLen+gcm.NonceSize()+len(plain)+gcm.Overhead())
	out = append(out, magic...)
	out = append(out, version, keyID)
	out = append(out, nonce...)
	return gcm.Seal(out, nonce, plain, nil), nil
}

func decrypt(blob, key []byte) (map[string]string, error) {
	if len(blob) < headerLen || string(blob[:len(magic)]) != string(magic) {
		return nil, fmt.Errorf("tiroir: %s is not a tiroir store", storeName)
	}
	if blob[len(magic)] != version {
		return nil, fmt.Errorf("tiroir: unsupported store version %d", blob[len(magic)])
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(blob) < headerLen+gcm.NonceSize() {
		return nil, fmt.Errorf("tiroir: store too short")
	}
	nonce := blob[headerLen : headerLen+gcm.NonceSize()]
	raw := blob[headerLen+gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, raw, nil)
	if err != nil {
		return nil, fmt.Errorf("tiroir: decrypt: %w", err)
	}
	vals := map[string]string{}
	if err := json.Unmarshal(plain, &vals); err != nil {
		return nil, fmt.Errorf("tiroir: corrupt payload: %w", err)
	}
	return vals, nil
}

func atomicWrite(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if err := tmp.Chmod(filePerm); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}