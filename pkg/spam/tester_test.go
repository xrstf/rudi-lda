// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package spam

import (
	"context"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"go.xrstf.de/rudi-lda/pkg/test"
	"go.xrstf.de/rudi-lda/pkg/test/emails"
)

func TestCheck(t *testing.T) {
	ctx := context.Background()
	msg := emails.GitHubIssueClosed()

	testcases := []struct {
		script   string
		expected *Result
	}{
		{
			// allow empty script files (normally Rudi programs cannot be empty)
			script:   ``,
			expected: nil,
		},
		{
			script:   `"an invalid return value"`,
			expected: nil,
		},
		{
			script: `(spam)`,
			expected: &Result{
				Status: Spam,
			},
		},
		{
			script: `(maybe-spam)`,
			expected: &Result{
				Status: MaybeSpam,
			},
		},
		{
			script: `(ham)`,
			expected: &Result{
				Status: Ham,
			},
		},
		{
			script: `(spam "foo")`,
			expected: &Result{
				Status: Spam,
				Rule:   "foo",
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.script, func(t *testing.T) {
			scriptFile, err := test.TempScript(testcase.script)
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(scriptFile)

			result, err := Check(ctx, scriptFile, msg)
			if err != nil {
				t.Fatalf("Failed to perform spam check: %v", err)
			}

			if !cmp.Equal(testcase.expected, result) {
				t.Fatalf("Expected %+v, got %+v", testcase.expected, result)
			}
		})
	}
}
