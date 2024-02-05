// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	rudilog "go.xrstf.de/rudi-lda/pkg/log"
)

func debugCommand(ctx context.Context, opt options) error {
	logger := rudilog.New("debug.log")
	logger.WithFields(optFields(opt)).Info("Options")
	logger.WithFields(envFields()).Info("Environment")

	stdin, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %w", err)
	}

	logger.WithField("stdin", string(stdin)).Info("stdin")

	return nil
}

func optFields(opt options) logrus.Fields {
	return logrus.Fields{
		"from":         opt.fromAddress,
		"dest":         opt.destAddress,
		"spamScript":   opt.spamScript,
		"folderScript": opt.folderScript,
		"maildir":      opt.maildir,
		"datadir":      opt.datadir,
		"rentablo":     opt.rentablo,
		"sunnyportal":  opt.sunnyportal,
	}
}

func envFields() logrus.Fields {
	fields := logrus.Fields{}

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		value := ""
		if len(parts) > 1 {
			value = parts[1]
		}

		fields[parts[0]] = value
	}

	return fields
}
