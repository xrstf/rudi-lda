// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package processor

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"

	"go.xrstf.de/rudi-lda/pkg/email"
	"go.xrstf.de/rudi-lda/pkg/metrics"
)

func Pipeline(ctx context.Context, logger logrus.FieldLogger, processors []Processor, msg *email.Message, metricsData *metrics.Metrics) (*email.Message, error) {
	for _, processor := range processors {
		consumed, newMsg, err := tryProcessor(ctx, logger, processor, msg, metricsData)
		if err != nil {
			// remember this error forever
			if newMsg == nil {
				newMsg = msg
			}

			headerName := fmt.Sprintf("X-Rudi-LDA-%s-Error", processor.Name())
			newMsg.Header[headerName] = []string{err.Error()}

			// make the next processor use the updated message
			msg = newMsg

			// failed processors can still claim that a message was consumed
			if consumed {
				return msg, err
			}

			// processor failed and mail is not consumed, so we continue with the next processor
			logger.WithField("processor", processor.Name()).WithError(err).Error("Processor failed")
			continue
		}

		// all good :)
		if consumed {
			return newMsg, nil
		}

		// not consumed, not errored => not matched, try the next processor
		msg = newMsg
	}

	return msg, nil
}

func tryProcessor(ctx context.Context, logger logrus.FieldLogger, proc Processor, msg *email.Message, metricsData *metrics.Metrics) (consumed bool, newMsg *email.Message, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("processor panicked: %v: %s", err, debug.Stack())
			consumed = false
		}
	}()

	return proc.Process(ctx, logger, msg, metricsData)
}
