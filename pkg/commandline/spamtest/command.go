// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package spamtest

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"

	"go.xrstf.de/rudi-lda/pkg/commandline/options"
	"go.xrstf.de/rudi-lda/pkg/email"
	"go.xrstf.de/rudi-lda/pkg/spam"
)

type Options struct {
	Common *options.CommonOptions

	SpamScript   string
	FolderScript string
}

func Command(commonOpt *options.CommonOptions) *cli.Command {
	opt := &Options{
		Common: commonOpt,
	}

	return &cli.Command{
		Name:            "spamtest",
		Usage:           "prints spam and folder script results on stdout",
		HideHelpCommand: true,
		Flags: []cli.Flag{
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
		},
		Action: func(ctx context.Context, _ *cli.Command) error {
			return deliverCommand(ctx, opt)
		},
	}
}

func deliverCommand(ctx context.Context, opt *Options) error {
	// read data from stdin
	rawMail, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read from stdin: %w", err)
	}

	// parse email
	msg, err := email.ParseMessage(rawMail)
	if err != nil {
		return fmt.Errorf("failed to parse mail body: %w", err)
	}

	// run the test
	result, err := spam.Check(ctx, opt.SpamScript, msg)
	if err != nil {
		return fmt.Errorf("failed to run spam check: %w", err)
	}

	if result == nil {
		fmt.Println("(no match)")
	} else {
		fmt.Printf("status: %s\nrule: %s\n", result.Status, result.Rule)
	}

	return nil
}
