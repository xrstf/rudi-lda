// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package antispam

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"go.xrstf.de/rudi-lda/pkg/email"
	"go.xrstf.de/rudi-lda/pkg/fs"
	"go.xrstf.de/rudi-lda/pkg/metrics"
	"go.xrstf.de/rudi-lda/pkg/spam"
)

type Proc struct {
	scriptFile string
	backupDir  string
}

func New(scriptFile string, backupDir string) *Proc {
	return &Proc{
		scriptFile: scriptFile,
		backupDir:  backupDir,
	}
}

func (*Proc) Name() string {
	return "antispam"
}

func (p *Proc) Process(ctx context.Context, logger logrus.FieldLogger, msg *email.Message, metrics *metrics.Metrics) (consumed bool, updated *email.Message, err error) {
	result, err := spam.Check(ctx, p.scriptFile, msg)
	if err != nil {
		return false, nil, err
	}

	if result == nil {
		return false, msg, nil
	}

	msg.Header["X-Rudi-LDA-Antispam"] = []string{
		fmt.Sprintf("status:%s,rule:%s", result.Status, result.Rule),
	}

	logger = logger.WithField("rule", result.Rule)

	if result.Status == spam.Spam {
		logger.Info("Dropping spam.")

		metrics.Discarded++
		metrics.SpamRules[result.Rule]++

		if p.backupDir != "" {
			if _, err := fs.WriteEmail(p.backupDir, msg); err != nil {
				logger.WithError(err).Error("Failed to backup spam e-mail: %w", err)
				// if we cannot backup spam, we must deliver it to the inbodx to prevent data loss
				return false, msg, nil
			}
		}

		return true, nil, nil
	}

	logger.WithField("status", result.Status).Debug("Passed spamtest.")

	return false, msg, nil
}
