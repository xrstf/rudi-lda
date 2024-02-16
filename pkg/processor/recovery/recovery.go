// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package recovery

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"go.xrstf.de/rudi-lda/pkg/email"
	"go.xrstf.de/rudi-lda/pkg/metrics"
	"go.xrstf.de/rudi-lda/pkg/util"
)

type Proc struct {
	dumpDir string
}

func New(dumpDir string) *Proc {
	return &Proc{
		dumpDir: dumpDir,
	}
}

func (p *Proc) Matches(_ context.Context, _ logrus.FieldLogger, _ *email.Message) (bool, error) {
	return true, nil
}

func (p *Proc) Process(_ context.Context, logger logrus.FieldLogger, msg *email.Message, _ *metrics.Metrics) error {
	logger.Info("Recovering unprocessable e-mail.")

	if err := os.MkdirAll(p.dumpDir, 0775); err != nil {
		log.Printf("Error: cannot create recovery directory: %v", err)
		return nil
	}

	filename := filepath.Join(p.dumpDir, util.Filename())

	if err := os.WriteFile(filename, msg.Raw(), 0600); err != nil {
		log.Printf("Error: failed to write file: %v", err)
	}

	return nil
}
