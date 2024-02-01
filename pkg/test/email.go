// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package test

import (
	"net/mail"

	"go.xrstf.de/rudi-lda/pkg/email"
	"go.xrstf.de/rudi-lda/pkg/util"
)

type Builder struct {
	body    string
	headers mail.Header
}

func NewMessageBuilder() *Builder {
	return &Builder{
		headers: mail.Header{},
	}
}

func (b *Builder) WithRawHeader(key, value string) *Builder {
	b.headers[key] = []string{value}

	return b
}

func (b *Builder) WithHeader(key, value string) *Builder {
	return b.WithRawHeader(key, util.EncodeQuotedPrintable(value))
}

func (b *Builder) WithSubject(subject string) *Builder {
	return b.WithHeader("Subject", subject)
}

func (b *Builder) WithFrom(from string) *Builder {
	return b.WithRawHeader("From", from)
}

func (b *Builder) WithTo(to string) *Builder {
	return b.WithRawHeader("To", to)
}

func (b *Builder) WithReplyTo(replyTo string) *Builder {
	return b.WithRawHeader("Reply-To", replyTo)
}

func (b *Builder) WithBody(body string) *Builder {
	b.body = body
	return b
}

func (b *Builder) Build() *email.Message {
	return &email.Message{
		Header: b.headers,
		Body:   b.body,
	}
}
