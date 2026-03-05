package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// RenameFile renames a file, handling the overwrite check.
func RenameFile(src, dst string, overwrite bool) error {
	if src == dst {
		return nil
	}

	if !overwrite {
		if _, err := os.Stat(dst); err == nil {
			return fmt.Errorf("destination already exists: %s", dst)
		}
	}

	err := os.Rename(src, dst)
	if err != nil {
		// Cross-device: fall back to copy + remove
		if copyErr := copyFile(src, dst); copyErr != nil {
			return fmt.Errorf("rename failed and copy fallback failed: %w", copyErr)
		}
		if removeErr := os.Remove(src); removeErr != nil {
			return fmt.Errorf("copy succeeded but removing source failed: %w", removeErr)
		}
	}
	return nil
}

// MoveFile moves a file to a destination, creating directories as needed.
func MoveFile(src, dst string, overwrite, alwaysMove, leaveSymlink bool) error {
	if src == dst {
		return nil
	}

	if !overwrite {
		if _, err := os.Stat(dst); err == nil {
			return fmt.Errorf("move destination already exists: %s", dst)
		}
	}

	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating destination directory: %w", err)
	}

	err := os.Rename(src, dst)
	if err != nil {
		// Cross-device move: copy then optionally remove source
		if copyErr := copyFile(src, dst); copyErr != nil {
			return fmt.Errorf("move failed: %w", copyErr)
		}
		if alwaysMove {
			if removeErr := os.Remove(src); removeErr != nil {
				return fmt.Errorf("removing source after move: %w", removeErr)
			}
		}
	}

	if leaveSymlink {
		// Create symlink at original location pointing to new location
		if err := os.Symlink(dst, src); err != nil {
			return fmt.Errorf("creating symlink: %w", err)
		}
	}

	return nil
}

// copyFile copies a file from src to dst preserving permissions.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	info, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
