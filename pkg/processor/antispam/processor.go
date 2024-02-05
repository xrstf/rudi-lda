// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package antispam

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"go.xrstf.de/rudi-lda/pkg/email"
	"go.xrstf.de/rudi-lda/pkg/metrics"
	"go.xrstf.de/rudi-lda/pkg/spam"
)

type Proc struct {
	scriptFile string
}

func New(scriptFile string) *Proc {
	return &Proc{
		scriptFile: scriptFile,
	}
}

func (p *Proc) Matches(ctx context.Context, logger logrus.FieldLogger, msg *email.Message) (bool, error) {
	result, err := spam.Check(ctx, p.scriptFile, msg)
	if err != nil {
		return false, fmt.Errorf("script failed: %w", err)
	}

	if result == nil {
		return false, nil
	}

	if result.Status == spam.Spam {
		logger.WithField("rule", result.Rule).Info("Dropping spam.")
		return true, nil
	}

	return false, nil
}

func (p *Proc) Process(_ context.Context, logger logrus.FieldLogger, _ *email.Message, _ *metrics.Metrics) error {
	// do nothing, just drop the message
	return nil
}
