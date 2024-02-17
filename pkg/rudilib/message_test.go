// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package rudilib

import (
	"context"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.xrstf.de/rudi"

	"go.xrstf.de/rudi-lda/pkg/test"
	"go.xrstf.de/rudi-lda/pkg/test/emails"
)

func TestProcessMessage(t *testing.T) {
	ctx := context.Background()
	msg := emails.GitHubIssueClosed()

	testcases := []struct {
		script     string
		extraVars  rudi.Variables
		extraFuncs rudi.Functions
		expected   any
	}{
		{
			// allow empty script files (normally Rudi programs cannot be empty)
			script:   ``,
			expected: nil,
		},
		{
			script:   `"whatever"`,
			expected: "whatever",
		},
		{
			script:   `.from.name`,
			expected: "SomeGithubUser",
		},
		{
			script:   `.from.address`,
			expected: "notifications@github.com",
		},
		{
			script:   `.to.name`,
			expected: "black7375/Firefox-UI-Fix",
		},
		{
			script:   `.to.address`,
			expected: "Firefox-UI-Fix@noreply.github.com",
		},
		{
			script:   `(domain .from.address)`,
			expected: "github.com",
		},
		{
			script:   `(domain .from)`,
			expected: "github.com",
		},
		{
			script:   `(user .from.address)`,
			expected: "notifications",
		},
		{
			script:   `(user .from)`,
			expected: "notifications",
		},
		{
			script:   `.subject`,
			expected: "Re: [black7375/Firefox-UI-Fix] Theme can't apply (Issue #868)",
		},
		{
			script:   `.headers.Precedence`,
			expected: []any{"list"},
		},
		{
			script:   `(header "DoesNotExist")`,
			expected: "",
		},
		{
			script:   `(header "Precedence" .)`,
			expected: "list",
		},
		{
			script:   `(header "Precedence")`,
			expected: "list",
		},
		{
			script:   `(header "preceDENCE")`,
			expected: "list",
		},
		{
			script:   `(header "Received")`,
			expected: `from [192.30.252.201] (out-18.smtp.github.com) by mailserver.example.com (chasquid) with ESMTPS tls TLS_AES_128_GCM_SHA256 (over SMTP, TLS-1.3, envelope from "noreply@github.com") ; Sat, 17 Feb 2024 15:20:18 +0000`,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.script, func(t *testing.T) {
			scriptFile, err := test.TempScript(testcase.script)
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(scriptFile)

			result, err := ProcessMessage(ctx, scriptFile, msg, testcase.extraVars, testcase.extraFuncs)
			if err != nil {
				t.Fatalf("Failed to process: %v", err)
			}

			if !cmp.Equal(testcase.expected, result) {
				t.Fatalf("Expected %+v, got %+v", testcase.expected, result)
			}
		})
	}
}
