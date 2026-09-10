package env

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := NewAt(dir)
	if err := s.Set("OPENROUTER_API_KEY", "sk-secret"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if err := s.Set("LOG_LEVEL", "debug"); err != nil {
		t.Fatalf("set2: %v", err)
	}
	got, ok, err := s.Get("OPENROUTER_API_KEY")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !ok || got != "sk-secret" {
		t.Fatalf("get = %q, %v; want %q, true", got, ok, "sk-secret")
	}
}

func TestStoreIsEncryptedAtRest(t *testing.T) {
	dir := t.TempDir()
	s := NewAt(dir)
	if err := s.Set("GH_TOKEN", "ghp-abc123"); err != nil {
		t.Fatalf("set: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, storeName))
	if err != nil {
		t.Fatalf("read store: %v", err)
	}
	if strings.Contains(string(raw), "ghp-abc123") {
		t.Fatalf("store leaks plaintext value")
	}
	key, err := os.ReadFile(filepath.Join(dir, keyName))
	if err != nil {
		t.Fatalf("read key: %v", err)
	}
	if len(key) != 32 {
		t.Fatalf("key length = %d; want 32", len(key))
	}
}

func TestPerms(t *testing.T) {
	dir := t.TempDir()
	s := NewAt(dir)
	if err := s.Set("A", "1"); err != nil {
		t.Fatalf("set: %v", err)
	}
	for _, name := range []string{storeName, keyName} {
		info, err := os.Stat(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("stat %s: %v", name, err)
		}
		if perm := info.Mode().Perm(); perm != filePerm {
			t.Fatalf("%s perm = %v; want %v", name, perm, filePerm)
		}
	}
}

func TestMissingStoreIsEmpty(t *testing.T) {
	dir := t.TempDir()
	s := NewAt(dir)
	keys, err := s.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(keys) != 0 {
		t.Fatalf("keys = %v; want none", keys)
	}
}

func TestDeleteAndList(t *testing.T) {
	dir := t.TempDir()
	s := NewAt(dir)
	if err := s.Set("B", "2"); err != nil {
		t.Fatalf("set b: %v", err)
	}
	if err := s.Set("A", "1"); err != nil {
		t.Fatalf("set a: %v", err)
	}
	if err := s.Delete("B"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	keys, err := s.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(keys) != 1 || keys[0] != "A" {
		t.Fatalf("keys = %v; want [A]", keys)
	}
	if _, ok, _ := s.Get("B"); ok {
		t.Fatalf("B still present after delete")
	}
}

func TestExportSorted(t *testing.T) {
	dir := t.TempDir()
	s := NewAt(dir)
	if err := s.Set("Z", "zz"); err != nil {
		t.Fatalf("set z: %v", err)
	}
	if err := s.Set("A", "it's"); err != nil {
		t.Fatalf("set a: %v", err)
	}
	lines, err := s.Export()
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	want := []string{"export A='it'\\''s'", "export Z='zz'"}
	if len(lines) != 2 || lines[0] != want[0] || lines[1] != want[1] {
		t.Fatalf("export = %v; want %v", lines, want)
	}
}

func TestTamperedStoreFailsToDecrypt(t *testing.T) {
	dir := t.TempDir()
	s := NewAt(dir)
	if err := s.Set("A", "1"); err != nil {
		t.Fatalf("set: %v", err)
	}
	path := filepath.Join(dir, storeName)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	raw[len(raw)-1] ^= 0xff
	if err := os.WriteFile(path, raw, filePerm); err != nil {
		t.Fatalf("rewrite: %v", err)
	}
	if _, _, err := s.Get("A"); err == nil {
		t.Fatalf("tampered store decrypted without error")
	}
}

func TestPersistenceAcrossStores(t *testing.T) {
	dir := t.TempDir()
	if err := NewAt(dir).Set("A", "1"); err != nil {
		t.Fatalf("set: %v", err)
	}
	got, ok, err := NewAt(dir).Get("A")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !ok || got != "1" {
		t.Fatalf("persisted get = %q, %v", got, ok)
	}
}