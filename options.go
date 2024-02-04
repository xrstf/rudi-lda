// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"os"

	"github.com/spf13/pflag"
)

type options struct {
	spamScript   string
	folderScript string
	maildir      string
	datadir      string
	rentablo     bool
	sunnyportal  bool
	version      bool
}

func (o *options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.spamScript, "spam-script", o.spamScript, "Rudi script that will be evaluated to determine if the incoming e-mail is spam ($RUDILDA_SPAM_SCRIPT).")
	fs.StringVar(&o.folderScript, "folder-script", o.folderScript, "Rudi script that will be evaluated to determine the target folder for an incoming e-mail ($RUDILDA_FOLDER_SCRIPT).")
	fs.StringVar(&o.maildir, "maildir", o.maildir, "Path to the root of the user's Maildir directory ($RUDILDA_MAILDIR).")
	fs.StringVar(&o.datadir, "datadir", o.datadir, "Path to where metrics and other data files should be placed ($RUDILDA_DATADIR).")
	pflag.BoolVar(&o.rentablo, "rentablo", o.rentablo, "Enable the rentablo.de processor ($RUDILDA_RENTABLO).")
	pflag.BoolVar(&o.sunnyportal, "sunnyportal", o.sunnyportal, "Enable the sunnyportal.de processor ($RUDILDA_SUNNYPORTAL).")
	pflag.BoolVarP(&o.version, "version", "V", o.version, "Show version info and exit immediately.")
}

func env(name string) string {
	return os.Getenv("RUDILDA_" + name)
}

func envEnabled(name string) bool {
	val := env(name)

	return val != "" && val != "false" && val != "0"
}

func (o *options) ApplyEnvironment() error {
	if o.spamScript == "" {
		o.spamScript = env("SPAM_SCRIPT")
	}

	if o.folderScript == "" {
		o.folderScript = env("FOLDER_SCRIPT")
	}

	if o.maildir == "" {
		o.maildir = env("MAILDIR")
	}

	if o.datadir == "" {
		o.datadir = env("DATADIR")
	}

	if !o.rentablo {
		o.rentablo = envEnabled("RENTABLO")
	}

	if !o.sunnyportal {
		o.sunnyportal = envEnabled("SUNNYPORTAL")
	}

	return nil
}

// Validate checks constraints that apply to all commands.
func (o *options) Validate() error {
	return nil
}
