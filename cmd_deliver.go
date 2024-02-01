// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"go.xrstf.de/rudi-lda/pkg/email"
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

	metricsData.Valid++

	if err := processMessage(ctx, opt, msg, metricsData); err != nil {
		log.Printf("E-mail is unprocessable: %v", err)
	}

	return nil
}

func processMessage(ctx context.Context, opt options, msg *email.Message, metricsData *metrics.Metrics) error {
	for _, processor := range getProcessors(opt) {
		var matches bool

		matches, err := processor.Matches(ctx, msg)
		if err != nil {
			log.Printf("Warning: Processor matched failed: %v", err)
			continue
		}

		if !matches {
			continue
		}

		err = processor.Process(ctx, msg, metricsData)
		if err != nil {
			log.Printf("Warning: Processor failed: %v", err)
		}
	}

	return nil
}

func getProcessors(opt options) []processor.Processor {
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
	processors = append(processors, maildir.New(opt.maildir, opt.folderScript))

	// in case any of the above fail, this one will dump the email for later debugging;
	// this processor never returns an error
	processors = append(processors, recovery.New(opt.datadir))

	return processors
}
