// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package antispam

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"go.xrstf.de/rudi-lda/pkg/email"
	"go.xrstf.de/rudi-lda/pkg/metrics"
	"go.xrstf.de/rudi-lda/pkg/spam"
	"go.xrstf.de/rudi-lda/pkg/util"
)

type Proc struct {
	scriptFile string
	spamDir    string

	matchedRule string
}

func New(scriptFile string, spamDir string) *Proc {
	return &Proc{
		scriptFile: scriptFile,
		spamDir:    spamDir,
	}
}

func (p *Proc) Matches(ctx context.Context, logger logrus.FieldLogger, msg *email.Message) (bool, error) {
	result, err := spam.Check(ctx, p.scriptFile, msg)
	if err != nil {
		return false, err
	}

	if result == nil {
		return false, nil
	}

	if result.Status == spam.Spam {
		p.matchedRule = result.Rule
		logger.WithField("rule", result.Rule).Info("Dropping spam.")
		return true, nil
	}

	return false, nil
}

func (p *Proc) Process(_ context.Context, logger logrus.FieldLogger, msg *email.Message, metrics *metrics.Metrics) error {
	metrics.Discarded++
	metrics.SpamRules[p.matchedRule]++

	if err := os.MkdirAll(p.spamDir, 0775); err != nil {
		log.Printf("Error: cannot create spam directory: %v", err)
		return nil
	}

	filename := filepath.Join(p.spamDir, util.Filename())

	if err := os.WriteFile(filename, msg.Raw(), 0600); err != nil {
		log.Printf("Error: failed to write file: %v", err)
	}

	return nil
}
