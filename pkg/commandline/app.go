// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package commandline

import (
	"fmt"
	"runtime"

	"github.com/urfave/cli/v3"

	"go.xrstf.de/rudi-lda/pkg/commandline/deliver"
	"go.xrstf.de/rudi-lda/pkg/commandline/options"
	"go.xrstf.de/rudi-lda/pkg/commandline/spamtest"
)

func newVersionPrinter(buildTag, buildCommit, buildDate string) func(cmd *cli.Command) {
	return func(cmd *cli.Command) {
		// handle empty values in case `go install` was used
		if buildCommit == "" {
			fmt.Printf("rudi-lda dev, built with %s\n",
				runtime.Version(),
			)
		} else {
			fmt.Printf("rudi-lda %s (%s), built with %s on %s\n",
				buildTag,
				buildCommit[:10],
				runtime.Version(),
				buildDate,
			)
		}
	}
}

func NewApp(buildTag, buildCommit, buildDate string) *cli.Command {
	opt := &options.CommonOptions{}

	cli.VersionPrinter = newVersionPrinter(buildTag, buildCommit, buildDate)

	// Having an empty Version would disable the --version flag,
	// which we do not want to just lose.
	version := buildTag
	if version == "" {
		version = "dev"
	}

	return &cli.Command{
		Name:    "rudi-lda",
		Usage:   "Filter e-mails with Rudi and deliver them to Maildirs++",
		Version: version,
		Commands: []*cli.Command{
			deliver.Command(opt),
			spamtest.Command(opt),
		},
	}
}
