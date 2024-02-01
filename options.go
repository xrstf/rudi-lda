// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
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
	fs.StringVar(&o.spamScript, "spam-script", o.spamScript, "Rudi script that will be evaluated to determine if the incoming e-mail is spam.")
	fs.StringVar(&o.folderScript, "folder-script", o.folderScript, "Rudi script that will be evaluated to determine the target folder for an incoming e-mail.")
	fs.StringVar(&o.maildir, "maildir", o.maildir, "Path to the root of the user's Maildir directory.")
	fs.StringVar(&o.datadir, "datadir", o.datadir, "Path to where metrics and other data files should be placed.")
	pflag.BoolVar(&o.rentablo, "rentablo", o.rentablo, "Enable the rentablo.de processor.")
	pflag.BoolVar(&o.sunnyportal, "sunnyportal", o.sunnyportal, "Enable the sunnyportal.de processor.")
	pflag.BoolVarP(&o.version, "version", "V", o.version, "Show version info and exit immediately.")
}

// Validate checks constraints that apply to all commands.
func (o *options) Validate() error {
	return nil
}
