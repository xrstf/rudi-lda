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
		metrics.SpamRules[p.matchedRule]++

		if _, err := util.WriteEmail(p.spamDir, msg); err != nil {
			logger.WithError(err).Error("Failed to backup spam e-mail: %w", err)
			// if we cannot backup spam, we must deliver it to the inbodx to prevent data loss
			return false, msg, nil
		}

		return true, nil, nil
	}

	logger.WithField("status", result.Status).Debug("Passed spamtest.")

	return false, msg, nil
}
