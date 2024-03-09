// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package email

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type Message struct {
	Header mail.Header
	Body   string
	raw    []byte
}

func ParseMessage(rawMessage []byte) (*Message, error) {
	msg, err := mail.ReadMessage(bytes.NewReader(rawMessage))
	if err != nil {
		return nil, fmt.Errorf("failed to parse mail body: %v", err)
	}

	body, _ := io.ReadAll(msg.Body)

	return &Message{
		Header: msg.Header,
		Body:   string(body),
		raw:    rawMessage,
	}, nil
}

func (m *Message) LogFields() logrus.Fields {
	fields := logrus.Fields{
		"subject": m.GetSubject(),
	}

	if from := m.GetFrom(); from != nil {
		fields["fromName"] = decodeQuotedPrintable(from.Name)
		fields["fromAddress"] = from.Address
	}

	if to := m.GetTo(); to != nil {
		fields["toName"] = decodeQuotedPrintable(to.Name)
		fields["toAddress"] = to.Address
	}

	return fields
}

func (m *Message) Raw() []byte {
	return m.raw
}

func (m *Message) GetDate() (time.Time, error) {
	date := m.Header.Get("Date")
	formats := []string{
		time.RFC822, time.RFC822Z,
		// time.RFC1123 but with _2 instead of 02
		"Mon, _2 Jan 2006 15:04:05 MST", "Mon, _2 Jan 2006 15:04:05 -0700",
		"Mon, _2 Jan 2006 15:04:05 MST (-0700)", "Mon, _2 Jan 2006 15:04:05 -0700 (MST)",
		"_2 Jan 06 15:04:05 MST", "_2 Jan 06 15:04:05 -0700",
		"_2 Jan 2006 15:04:05 MST", "_2 Jan 2006 15:04:05 -0700",
	}

	for _, format := range formats {
		parsed, err := time.Parse(format, date)
		if err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("cannot parse %q as a time", date)
}

func (m *Message) GetSubject() string {
	return decodeQuotedPrintable(m.Header.Get("Subject"))
}

func (m *Message) GetFrom() *mail.Address {
	return m.getAddress("From")
}

func (m *Message) GetTo() *mail.Address {
	return m.getAddress("To")
}

func (m *Message) GetReplyTo() *mail.Address {
	return m.getAddress("Reply-To")
}

func (m *Message) GetDeliveredTo() string {
	return m.Header.Get("Delivered-To")
}

func (m *Message) getAddress(header string) *mail.Address {
	list, err := m.Header.AddressList(header)
	if err != nil || len(list) == 0 {
		return nil
	}

	return list[0]
}

// based on https://github.com/kirabou/parseMIMEemail.go/blob/master/parseMIMEmail.go
func (m *Message) GetMultipartBody(contentType string) (string, error) {
	mediaType, params, err := mime.ParseMediaType(m.Header.Get("Content-Type"))
	if err != nil {
		return "", fmt.Errorf("failed to parse media type: %w", err)
	}

	if !strings.HasPrefix(mediaType, "multipart/") {
		return "", nil
	}

	// Instantiate a new io.Reader dedicated to MIME multipart parsing
	// using multipart.NewReader()
	reader := multipart.NewReader(strings.NewReader(m.Body), params["boundary"])
	if reader == nil {
		return "", nil
	}

	// Go through each of the MIME part of the message Body with NextPart(),
	// and read the content of the MIME part with ioutil.ReadAll()
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed going through the MIME parts: %w", err)
		}

		mediaType, _, err := mime.ParseMediaType(part.Header.Get("Content-Type"))
		if err != nil {
			return "", fmt.Errorf("failed to parse MIME headers: %w", err)
		}

		// not the content type we are looking for
		if mediaType != contentType {
			continue
		}

		rawPartBody, err := io.ReadAll(part)
		if err != nil {
			return "", fmt.Errorf("failed to read MIME body: %w", err)
		}

		encoding := strings.ToUpper(part.Header.Get("Content-Transfer-Encoding"))
		body := string(rawPartBody)

		switch {
		case encoding == "BASE64":
			decoded, err := base64.StdEncoding.DecodeString(body)
			if err != nil {
				return "", fmt.Errorf("failed to base64 decode MIME body: %w", err)
			}
			body = string(decoded)

		case encoding == "QUOTED-PRINTABLE":
			decoded, err := io.ReadAll(quotedprintable.NewReader(strings.NewReader(body)))
			if err != nil {
				return "", fmt.Errorf("failed to decode quoted-printable MIME body: %w", err)
			}
			body = string(decoded)
		}

		return body, nil
	}

	return "", nil
}

func decodeQuotedPrintable(s string) string {
	dec := new(mime.WordDecoder)
	b, _ := dec.DecodeHeader(s)

	return string(b)
}
