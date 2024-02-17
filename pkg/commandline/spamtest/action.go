// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package spamtest

import (
	"context"
	"fmt"
	"io"
	"os"

	"go.xrstf.de/rudi-lda/pkg/email"
	"go.xrstf.de/rudi-lda/pkg/spam"
)

func action(ctx context.Context, opt *Options) error {
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
