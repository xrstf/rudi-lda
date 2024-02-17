// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.xrstf.de/rudi-lda/pkg/email"
)

var (
	FilePermissions      os.FileMode = 0660
	DirectoryPermissions os.FileMode = 0770
)

func UniqueEmailFilename() string {
	now := time.Now().UTC().Format("20060102_150405")

	return fmt.Sprintf("%s_%d.eml", now, os.Getpid())
}

func WriteEmail(directory string, msg *email.Message) (string, error) {
	if err := os.MkdirAll(directory, DirectoryPermissions); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	filename := filepath.Join(directory, UniqueEmailFilename())

	if err := os.WriteFile(filename, msg.Raw(), FilePermissions); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filename, nil
}
