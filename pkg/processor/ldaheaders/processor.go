// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package ldaheaders

import (
	"context"

	"github.com/sirupsen/logrus"

	"go.xrstf.de/rudi-lda/pkg/email"
	"go.xrstf.de/rudi-lda/pkg/metrics"
)

type Proc struct {
	destUser string
}

func New(destUser string) *Proc {
	return &Proc{
		destUser: destUser,
	}
}

func (*Proc) Name() string {
	return "ldaheaders"
}

func (p *Proc) Process(_ context.Context, _ logrus.FieldLogger, msg *email.Message, _ *metrics.Metrics) (consumed bool, updated *email.Message, err error) {
	msg.Header["Delivered-To"] = []string{p.destUser}

	return false, msg, nil
}
