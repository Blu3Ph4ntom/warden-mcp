package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func WriteFileAtomic(path string, content []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".warden-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	if _, err := tmp.Write(content); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err == nil {
		return nil
	} else if runtime.GOOS != "windows" {
		return err
	}
	backupPath := path + ".bak"
	_ = os.Remove(backupPath)
	if _, statErr := os.Stat(path); statErr == nil {
		if err := os.Rename(path, backupPath); err != nil {
			return fmt.Errorf("prepare windows replace: %w", err)
		}
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Rename(backupPath, path)
		return fmt.Errorf("replace file: %w", err)
	}
	_ = os.Remove(backupPath)
	return nil
}
