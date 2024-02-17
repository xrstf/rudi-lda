// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package deliver

import (
	"context"

	"github.com/urfave/cli/v3"

	"go.xrstf.de/rudi-lda/pkg/commandline/options"
)

type Options struct {
	Common *options.CommonOptions

	FromAddress  string
	DestAddress  string
	SpamScript   string
	FolderScript string
	BackupSpam   bool
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
			&cli.BoolFlag{
				Name:        "backup-spam",
				Usage:       "write spam e-mails to $datadir/spam",
				Sources:     cli.EnvVars("RUDILDA_BACKUP_SPAM"),
				Destination: &opt.BackupSpam,
			},
		},
		Action: func(ctx context.Context, _ *cli.Command) error {
			return action(ctx, opt)
		},
	}
}
