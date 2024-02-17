// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package maildir

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"go.xrstf.de/rudi-lda/pkg/email"
	"go.xrstf.de/rudi-lda/pkg/maildir"
	"go.xrstf.de/rudi-lda/pkg/metrics"
	"go.xrstf.de/rudi-lda/pkg/rudilib"
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

func (*Proc) Name() string {
	return "maildir"
}

func (p *Proc) Process(ctx context.Context, logger logrus.FieldLogger, msg *email.Message, metricsData *metrics.Metrics) (consumed bool, updated *email.Message, err error) {
	md, err := maildir.New(p.mailDirectory)
	if err != nil {
		return false, nil, fmt.Errorf("invalid maildir %q: %w", p.mailDirectory, err)
	}

	folder, err := p.determineFolder(ctx, msg)
	if err != nil {
		logger.WithError(err).Error("Failed to determine folder.")
		// continue, i.e. deliver into root maildir folder (inbox)
	}

	if err := md.Deliver(folder, msg); err != nil {
		return false, nil, fmt.Errorf("failed to deliver into maildir: %w", err)
	}

	return true, nil, nil
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
