// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package maildir

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go.xrstf.de/rudi-lda/pkg/email"
	"go.xrstf.de/rudi-lda/pkg/metrics"
	"go.xrstf.de/rudi-lda/pkg/rudilib"
	"go.xrstf.de/rudi-lda/pkg/util"
)

type Proc struct {
	mailDirectory string
	folderScript  string
}

func New(mailDirectory string, folderScript string) *Proc {
	return &Proc{
		mailDirectory: mailDirectory,
		folderScript:  folderScript,
	}
}

func (p *Proc) Matches(_ context.Context, _ *email.Message) (bool, error) {
	return true, nil
}

func (p *Proc) Process(ctx context.Context, msg *email.Message, metricsData *metrics.Metrics) error {
	folder, err := p.determineFolder(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to determine destination folder: %w", err)
	}

	destinationDir := p.mailDirectory
	if folder != "" {
		destinationDir = filepath.Join(destinationDir, "."+folder)
	}

	metricsData.Folders[folder]++

	tmpDir := filepath.Join(destinationDir, "tmp")
	if err := os.MkdirAll(tmpDir, 0700); err != nil {
		return fmt.Errorf("failed to ensure temp directory: %w", err)
	}

	filename := util.Filename()
	tmpFile := filepath.Join(tmpDir, filename)

	if err := os.WriteFile(tmpFile, msg.Raw(), 0600); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	newDir := filepath.Join(destinationDir, "new")
	if err := os.MkdirAll(newDir, 0700); err != nil {
		return fmt.Errorf("failed to ensure new directory: %w", err)
	}

	newFile := filepath.Join(newDir, filename)

	if err := os.Rename(tmpFile, newFile); err != nil {
		return fmt.Errorf("failed to move message to new directory: %w", err)
	}

	// Maildir folders need to be marked
	if folder != "" {
		markerFile := filepath.Join(destinationDir, "maildirfolder")

		if _, err := os.Stat(markerFile); err != nil {
			if err := os.WriteFile(markerFile, nil, 0666); err != nil {
				return fmt.Errorf("failed to mark: %w", err)
			}
		}
	}

	return nil
}

func (p *Proc) determineFolder(ctx context.Context, msg *email.Message) (string, error) {
	if p.folderScript == "" {
		return "", nil
	}

	result, err := rudilib.ProcessMessage(ctx, p.folderScript, msg, nil, nil)
	if err != nil {
		return "", fmt.Errorf("script failed: %w", err)
	}

	if result == nil {
		return "", nil
	}

	if s, ok := result.(string); ok {
		return s, nil
	}

	return "", fmt.Errorf("script did not return string, but %T", result)
}
