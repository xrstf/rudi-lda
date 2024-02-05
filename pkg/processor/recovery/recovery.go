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
	datadir string
}

func New(datadir string) *Proc {
	return &Proc{
		datadir: datadir,
	}
}

func (p *Proc) Matches(_ context.Context, _ logrus.FieldLogger, _ *email.Message) (bool, error) {
	return true, nil
}

func (p *Proc) Process(_ context.Context, logger logrus.FieldLogger, msg *email.Message, _ *metrics.Metrics) error {
	logger.Info("Recovering unprocessable e-mail.")

	directory := filepath.Join(p.datadir, "unprocessable")
	if err := os.MkdirAll(directory, 0755); err != nil {
		log.Printf("Error: cannot create recovery directory: %v", err)
		return nil
	}

	filename := filepath.Join(directory, util.Filename())

	if err := os.WriteFile(filename, msg.Raw(), 0600); err != nil {
		log.Printf("Error: failed to write file: %v", err)
	}

	return nil
}
