// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package processor

import (
	"context"

	"github.com/sirupsen/logrus"

	"go.xrstf.de/rudi-lda/pkg/email"
	"go.xrstf.de/rudi-lda/pkg/metrics"
)

type Processor interface {
	Name() string
	Process(ctx context.Context, logger logrus.FieldLogger, msg *email.Message, metricsData *metrics.Metrics) (consumed bool, updated *email.Message, err error)
}
