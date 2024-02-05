// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"go.xrstf.de/rudi-lda/pkg/email"
	"go.xrstf.de/rudi-lda/pkg/log"
	"go.xrstf.de/rudi-lda/pkg/metrics"
	"go.xrstf.de/rudi-lda/pkg/processor"
	"go.xrstf.de/rudi-lda/pkg/processor/antispam"
	"go.xrstf.de/rudi-lda/pkg/processor/maildir"
	"go.xrstf.de/rudi-lda/pkg/processor/recovery"
	"go.xrstf.de/rudi-lda/pkg/processor/rentablo"
	"go.xrstf.de/rudi-lda/pkg/processor/sunnyportal"
)

func deliverCommand(ctx context.Context, opt options) error {
	if opt.maildir == "" {
		return errors.New("--maildir must be configured")
	}

	if opt.datadir == "" {
		return errors.New("--datadir must be configured")
	}

	var metricsData *metrics.Metrics

	// read data from stdin
	rawMail, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read from stdin: %w", err)
	}

	// init metrics
	if opt.datadir != "" {
		metricsFile := filepath.Join(opt.datadir, "metrics.json")

		metricsData, err = metrics.Load(metricsFile)
		if err != nil {
			return fmt.Errorf("failed to load metrics: %w", err)
		}

		metricsData.Total++
		defer metrics.Save(metricsFile, metricsData)
	} else {
		metricsData = &metrics.Metrics{}
	}

	// parse email
	msg, err := email.ParseMessage(rawMail)
	if err != nil {
		return fmt.Errorf("failed to parse mail body: %w", err)
	}

	logger := log.New("mails.log").WithFields(msg.LogFields()).WithField("destination", opt.destAddress)

	metricsData.Valid++

	if err := processMessage(ctx, logger, opt, msg, metricsData); err != nil {
		logger.WithError(err).Warn("E-mail is unprocessable")
	}

	return nil
}

func processMessage(ctx context.Context, logger logrus.FieldLogger, opt options, msg *email.Message, metricsData *metrics.Metrics) error {
	for _, processor := range getProcessors(opt) {
		var matches bool

		matches, err := processor.Matches(ctx, logger, msg)
		if err != nil {
			logger.WithError(err).Warn("Processor matched failed")
			continue
		}

		if !matches {
			continue
		}

		err = processor.Process(ctx, logger, msg, metricsData)
		if err != nil {
			logger.WithError(err).Warn("Processor failed")
		} else {
			break
		}
	}

	return nil
}

func getProcessors(opt options) []processor.Processor {
	// assemble the path to the destination user's maildir
	userMaildir := getDestinationMaildir(opt)

	var processors []processor.Processor

	if opt.rentablo {
		processors = append(processors, rentablo.New(opt.datadir))
	}

	if opt.sunnyportal {
		processors = append(processors, sunnyportal.New(opt.datadir))
	}

	if opt.spamScript != "" {
		processors = append(processors, antispam.New(opt.spamScript))
	}

	// maildir will always match any e-mail
	processors = append(processors, maildir.New(userMaildir, opt.folderScript))

	// in case any of the above fail, this one will dump the email for later debugging;
	// this processor never returns an error
	processors = append(processors, recovery.New(opt.datadir))

	return processors
}

func getDestinationMaildir(opt options) string {
	parts := strings.Split(opt.destAddress, "@")

	return filepath.Join(opt.maildir, parts[0])
}
