// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package maildir

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"go.xrstf.de/rudi-lda/pkg/email"
	"go.xrstf.de/rudi-lda/pkg/fs"
)

type Maildir struct {
	baseDir string
}

func New(baseDir string) (*Maildir, error) {
	info, err := os.Stat(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to stat maildir: %w", err)
	}
	if !info.IsDir() {
		return nil, errors.New("maildir is not a directory")
	}

	return &Maildir{
		baseDir: baseDir,
	}, nil
}

func (m *Maildir) Deliver(folder string, msg *email.Message) error {
	destinationDir := m.baseDir
	if folder != "" {
		destinationDir = filepath.Join(destinationDir, "."+folder)
	}

	// create temporary file first
	tmpFile, err := fs.WriteEmail(filepath.Join(destinationDir, "tmp"), msg)
	if err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// ensure ./new exists
	newDir := filepath.Join(destinationDir, "new")
	if err := os.MkdirAll(newDir, fs.DirectoryPermissions); err != nil {
		return fmt.Errorf("failed to ensure new directory: %w", err)
	}

	// move file atomically to the new directory
	newFile := filepath.Join(newDir, filepath.Base(tmpFile))
	if err := os.Rename(tmpFile, newFile); err != nil {
		return fmt.Errorf("failed to move message to new directory: %w", err)
	}

	// Maildir folders need to be marked
	if folder != "" {
		markerFile := filepath.Join(destinationDir, "maildirfolder")

		if _, err := os.Stat(markerFile); err != nil {
			if err := os.WriteFile(markerFile, nil, fs.FilePermissions); err != nil {
				return fmt.Errorf("failed to mark: %w", err)
			}
		}
	}

	return nil
}
