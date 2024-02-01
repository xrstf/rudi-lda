// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package antispam

import (
	"context"
	"fmt"

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

func (p *Proc) Matches(ctx context.Context, msg *email.Message) (bool, error) {
	result, err := spam.Check(ctx, p.scriptFile, msg)
	if err != nil {
		return false, fmt.Errorf("script failed: %w", err)
	}

	return result.Status == spam.Spam, nil
}

func (p *Proc) Process(_ context.Context, _ *email.Message, _ *metrics.Metrics) error {
	// do nothing, just drop the message
	return nil
}
