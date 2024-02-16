// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package deliver

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"

	"go.xrstf.de/rudi-lda/pkg/commandline/options"
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

type Options struct {
	Common *options.CommonOptions

	FromAddress  string
	DestAddress  string
	SpamScript   string
	FolderScript string
	MailDir      string
	DataDir      string
	Rentablo     bool
	Sunnyportal  bool
}

func Command(commonOpt *options.CommonOptions) *cli.Command {
	opt := &Options{
		Common: commonOpt,
	}

	return &cli.Command{
		Name:            "deliver",
		Usage:           "delivers e-mail into a Maildir++ folder (default command)",
		HideHelpCommand: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "maildir",
				Usage:       "(required) path to the root of the user's Maildir directory",
				Sources:     cli.EnvVars("RUDILDA_MAILDIR"),
				Destination: &opt.MailDir,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "datadir",
				Usage:       "(required) path to where metrics and other data files should be placed",
				Sources:     cli.EnvVars("RUDILDA_DATADIR"),
				Destination: &opt.DataDir,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "from",
				Aliases:     []string{"f"},
				Usage:       "from address",
				Destination: &opt.FromAddress,
			},
			&cli.StringFlag{
				Name:        "destination",
				Aliases:     []string{"d"},
				Usage:       "(required) destination address",
				Destination: &opt.DestAddress,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "spam-script",
				Usage:       "Rudi script that will be evaluated to determine if the incoming e-mail is spam",
				Sources:     cli.EnvVars("RUDILDA_SPAM_SCRIPT"),
				Destination: &opt.SpamScript,
			},
			&cli.StringFlag{
				Name:        "folder-script",
				Usage:       "Rudi script that will be evaluated to determine the target folder for an incoming e-mail",
				Sources:     cli.EnvVars("RUDILDA_FOLDER_SCRIPT"),
				Destination: &opt.FolderScript,
			},
			&cli.BoolFlag{
				Name:        "rentablo",
				Usage:       "enable the rentablo.de processor",
				Sources:     cli.EnvVars("RUDILDA_RENTABLO"),
				Destination: &opt.Rentablo,
			},
			&cli.BoolFlag{
				Name:        "sunnyportal",
				Usage:       "enable the sunnyportal.de processor",
				Sources:     cli.EnvVars("RUDILDA_SUNNYPORTAL"),
				Destination: &opt.Sunnyportal,
			},
		},
		Action: func(ctx context.Context, _ *cli.Command) error {
			return deliverCommand(ctx, opt)
		},
	}
}

func deliverCommand(ctx context.Context, opt *Options) error {
	if err := log.SetDirectory(opt.DataDir); err != nil {
		return fmt.Errorf("invalid --datadir: %w", err)
	}

	var metricsData *metrics.Metrics

	// read data from stdin
	rawMail, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read from stdin: %w", err)
	}

	// init metrics
	if opt.DataDir != "" {
		metricsFile := filepath.Join(opt.DataDir, "metrics.json")

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

	logger := log.New("mails.log").WithFields(msg.LogFields()).WithField("destination", opt.DestAddress)

	metricsData.Valid++

	if err := processMessage(ctx, logger, opt, msg, metricsData); err != nil {
		logger.WithError(err).Warn("E-mail is unprocessable")
	}

	return nil
}

func processMessage(ctx context.Context, logger logrus.FieldLogger, opt *Options, msg *email.Message, metricsData *metrics.Metrics) error {
	for _, processor := range getProcessors(opt) {
		done, err := tryProcessor(ctx, logger, processor, msg, metricsData)
		if err != nil {
			logger.WithError(err).Warn("Processor failed")
			continue
		}

		if done {
			break
		}
	}

	return nil
}

func tryProcessor(ctx context.Context, logger logrus.FieldLogger, proc processor.Processor, msg *email.Message, metricsData *metrics.Metrics) (done bool, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("processor panicked: %v: %s", err, debug.Stack())
			done = false
		}
	}()

	matches, err := proc.Matches(ctx, logger, msg)
	if err != nil {
		return false, fmt.Errorf("matching failed: %w", err)
	}

	if !matches {
		return false, nil
	}

	err = proc.Process(ctx, logger, msg, metricsData)
	if err != nil {
		return false, fmt.Errorf("processor failed: %w", err)
	}

	return true, nil
}

func getProcessors(opt *Options) []processor.Processor {
	// assemble the path to the destination user's maildir
	userMaildir := getDestinationMaildir(opt)

	var processors []processor.Processor

	if opt.Rentablo {
		processors = append(processors, rentablo.New(opt.DataDir))
	}

	if opt.Sunnyportal {
		processors = append(processors, sunnyportal.New(opt.DataDir))
	}

	if opt.SpamScript != "" {
		processors = append(processors, antispam.New(opt.SpamScript, filepath.Join(opt.DataDir, "spam")))
	}

	// maildir will always match any e-mail
	processors = append(processors, maildir.New(userMaildir, opt.FolderScript))

	// in case any of the above fail, this one will dump the email for later debugging;
	// this processor never returns an error
	processors = append(processors, recovery.New(filepath.Join(opt.DataDir, "unprocessable")))

	return processors
}

func getDestinationMaildir(opt *Options) string {
	parts := strings.Split(opt.DestAddress, "@")

	return filepath.Join(opt.MailDir, parts[0])
}
