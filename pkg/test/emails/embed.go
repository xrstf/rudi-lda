// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package emails

import (
	_ "embed"

	"go.xrstf.de/rudi-lda/pkg/email"
)

//go:embed patreon-update.eml
var PatreonUpdateRaw []byte

func PatreonUpdate() *email.Message {
	return parse(PatreonUpdateRaw)
}

//go:embed github-issue-closed.eml
var GitHubIssueClosedRaw []byte

func GitHubIssueClosed() *email.Message {
	return parse(GitHubIssueClosedRaw)
}

func parse(content []byte) *email.Message {
	msg, err := email.ParseMessage(content)
	if err != nil {
		panic(err)
	}

	return msg
}
