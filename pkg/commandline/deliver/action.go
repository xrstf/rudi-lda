// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package deliver

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"go.xrstf.de/rudi-lda/pkg/email"
	"go.xrstf.de/rudi-lda/pkg/fs"
	"go.xrstf.de/rudi-lda/pkg/log"
	"go.xrstf.de/rudi-lda/pkg/metrics"
	"go.xrstf.de/rudi-lda/pkg/processor"
	"go.xrstf.de/rudi-lda/pkg/processor/antispam"
	"go.xrstf.de/rudi-lda/pkg/processor/ldaheaders"
	"go.xrstf.de/rudi-lda/pkg/processor/maildir"
	"go.xrstf.de/rudi-lda/pkg/processor/rentablo"
	"go.xrstf.de/rudi-lda/pkg/processor/sunnyportal"
)

func action(ctx context.Context, opt *Options) error {
	if err := log.SetDirectory(opt.DataDir); err != nil {
		return fmt.Errorf("invalid --datadir: %w", err)
	}

	// read data from stdin
	rawMail, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read from stdin: %w", err)
	}

	// setup logger
	var logger logrus.FieldLogger = log.New("mails.log")

	// init metrics
	metricsFile := filepath.Join(opt.DataDir, "metrics.json")
	metricsData, err := metrics.Load(metricsFile)
	if err != nil {
		return fmt.Errorf("failed to load metrics: %w", err)
	}

	defer func() {
		if err := metrics.Save(metricsFile, metricsData); err != nil {
			logger.WithError(err).Error("Failed to save metrics.")
		}
	}()

	metricsData.Total++

	// parse email
	msg, err := email.ParseMessage(rawMail)
	if err != nil {
		return fmt.Errorf("failed to parse mail body: %w", err)
	}

	metricsData.Valid++

	// process it
	logger = logger.WithFields(msg.LogFields()).WithField("destination", opt.DestUser)
	processors := getProcessors(opt)

	if newMsg, err := processor.Pipeline(ctx, logger, processors, msg, metricsData); err != nil {
		logger.WithError(err).Error("E-mail is unprocessable")

		// try to backup the e-mail for further debugging
		if _, err := fs.WriteEmail(filepath.Join(opt.DataDir, "unprocessable"), newMsg); err != nil {
			logger.WithError(err).Error("Failed to backup e-mail, too.")
		}
	}

	return nil
}

func getProcessors(opt *Options) []processor.Processor {
	// assemble the path to the destination user's maildir
	userMaildir := getDestinationMaildir(opt)

	var processors []processor.Processor

	// add common headers
	processors = append(processors, ldaheaders.New(opt.DestUser))

	if opt.Rentablo {
		processors = append(processors, rentablo.New(opt.DataDir))
	}

	if opt.Sunnyportal {
		processors = append(processors, sunnyportal.New(opt.DataDir))
	}

	if opt.SpamScript != "" {
		var backupDir string
		if opt.BackupSpam {
			backupDir = filepath.Join(opt.DataDir, "spam")
		}

		processors = append(processors, antispam.New(opt.SpamScript, backupDir))
	}

	// maildir will always consume any e-mail
	processors = append(processors, maildir.New(userMaildir, opt.FolderScript))

	return processors
}

func getDestinationMaildir(opt *Options) string {
	parts := strings.Split(opt.DestUser, "@")

	return filepath.Join(opt.MailDir, parts[0])
}
