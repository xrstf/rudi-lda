// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"slices"
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
		log.Fatal("No command given, use one of deliver, spamtest, debug.")
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
	case "debug":
		fmt.Println("Options:")
		fmt.Printf("  fromAddress: %q\n", opt.fromAddress)
		fmt.Printf("  destAddress: %q\n", opt.destAddress)
		fmt.Printf("  spamScript: %q\n", opt.spamScript)
		fmt.Printf("  folderScript: %q\n", opt.folderScript)
		fmt.Printf("  maildir: %q\n", opt.maildir)
		fmt.Printf("  datadir: %q\n", opt.datadir)
		fmt.Printf("  rentablo: %v\n", opt.rentablo)
		fmt.Printf("  sunnyportal: %v\n", opt.sunnyportal)
		fmt.Println("Environment:")

		env := os.Environ()
		slices.Sort(env)

		for _, e := range env {
			fmt.Printf("  %s\n", e)
		}
		fmt.Println("stdin:")
		io.Copy(os.Stdout, os.Stdin)
	default:
		err = fmt.Errorf("unknown command %q", command)
	}

	if err != nil {
		log.Fatalf("Program failed: %v", err)
	}
}
