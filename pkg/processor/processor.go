// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package processor

import (
	"context"

	"go.xrstf.de/rudi-lda/pkg/email"
	"go.xrstf.de/rudi-lda/pkg/metrics"
)

type Processor interface {
	Matches(ctx context.Context, msg *email.Message) (bool, error)
	Process(ctx context.Context, msg *email.Message, metricsData *metrics.Metrics) error
}
