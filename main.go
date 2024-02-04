// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"strings"

	"github.com/spf13/pflag"
)

// These variables get set by ldflags during compilation.
var (
	BuildTag    string
	BuildCommit string
	BuildDate   string // RFC3339 format ("2006-01-02T15:04:05Z07:00")
)

func printVersion() {
	// handle empty values in case `go install` was used
	if BuildCommit == "" {
		fmt.Printf("rudi-lda dev, built with %s\n",
			runtime.Version(),
		)
	} else {
		fmt.Printf("rudi-lda %s (%s), built with %s on %s\n",
			BuildTag,
			BuildCommit[:10],
			runtime.Version(),
			BuildDate,
		)
	}
}

func main() {
	opt := options{}
	opt.AddFlags(pflag.CommandLine)
	pflag.Parse()

	if opt.version {
		printVersion()
		return
	}

	if pflag.NArg() == 0 {
		log.Fatal("No command given, use one of deliver, spamtest.")
	}

	err := opt.Validate()
	if err != nil {
		log.Fatalf("Invalid command line: %v", err)
	}

	command := strings.ToLower(pflag.Arg(0))

	switch command {
	case "deliver":
		err = deliverCommand(context.Background(), opt)
	case "spamtest":
		err = spamtestCommand(context.Background(), opt)
	default:
		err = fmt.Errorf("unknown command %q", command)
	}

	if err != nil {
		log.Fatalf("Program failed: %v", err)
	}
}
