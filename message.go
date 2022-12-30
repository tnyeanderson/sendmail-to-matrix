package main

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/mail"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

func buildPreface() (p string) {
	p = config["preface"]
	if p != "" {
		p = p + "\n"
	}
	return
}

func buildSubject(m *mail.Message) (s string) {
	s = m.Header.Get("Subject")
	if s != "" {
		s = "Subject: " + s + "\n"
	}
	return
}

func buildBody(m *mail.Message) string {
	messageType, params := getMessageType(m)
	if strings.HasPrefix(messageType, "multipart/") {
		boundary := params["boundary"]
		if boundary != "" {
			content, err := parseMultipart(m, messageType, boundary)
			if err == nil {
				return string(content)
			}
		}
	}
	b, err := io.ReadAll(m.Body)
	if err != nil {
		log.Fatal(err)
	}
	return string(b)
}

func buildMessage(email io.Reader) (message string) {
	m, err := mail.ReadMessage(email)
	if err != nil {
		log.Fatal(err)
	}
	message += buildPreface()
	message += buildSubject(m)
	message += buildBody(m)
	return
}

func removeHtmlTags(message string) (s string) {
	policy := bluemonday.StrictPolicy()
	s = policy.Sanitize(message)
	s = html.UnescapeString(s)
	s = fixWhitespace(s)
	return
}

func fixWhitespace(message string) string {
	re := regexp.MustCompile("\n\n+")
	return re.ReplaceAllLiteralString(message, "\n")
}

func parseMultipart(m *mail.Message, messageType, boundary string) ([]byte, error) {
	mr := multipart.NewReader(m.Body, boundary)
	if messageType == "multipart/alternative" {
		return readAlternativeParts(mr)
	}
	if messageType == "multipart/mixed" {
		return readMixedParts(mr)
	}
	return nil, fmt.Errorf("Not a recognized multipart message")
}

// Only text/plain and text/html are recognized. In defiance of MIME (RFC2046),
// text/plain is preferred.
func readAlternativeParts(r *multipart.Reader) ([]byte, error) {
	parts := map[string][]byte{}
	for {
		p, err := r.NextPart()
		if err != nil {
			break
		}
		if yes, _ := partIsAttachment(p); yes {
			// Ignore attachments
			continue
		}
		partType := getPartType(p)
		bytes, err := io.ReadAll(p)
		if err == nil {
			// The last recognized result (accounting for text/plain preference) should
			// be returned. Using a map overrides previous parts with the same type.
			parts[partType] = bytes
		}
	}
	if bytes, ok := parts["text/plain"]; ok {
		return bytes, nil
	}
	if bytes, ok := parts["text/html"]; ok {
		return bytes, nil
	}
	return nil, fmt.Errorf("Unsupported alternative part")
}

func readMixedParts(r *multipart.Reader) ([]byte, error) {
	out := bytes.NewBuffer([]byte{})
	for {
		p, err := r.NextPart()
		if err != nil {
			break
		}
		if yes, _ := partIsAttachment(p); yes {
			// Ignore attachments for now
			// TODO: Handle attachments
			continue
		}
		partType := getPartType(p)
		// Only text parts are recognized.
		if strings.HasPrefix(partType, "text/") {
			// Ignore errors
			io.Copy(out, p)
		}
	}
	return out.Bytes(), nil
}

// Get the top-level media type and parameters. If not set, use the default
// according to RFC2045 5.2
func getMessageType(m *mail.Message) (contentType string, params map[string]string) {
	c := m.Header["Content-Type"]
	if len(c) > 0 {
		t, p, _ := mime.ParseMediaType(c[0])
		return t, p
	}
	return "text/plain", map[string]string{"charset": "us-ascii"}
}

func getPartType(p *multipart.Part) string {
	c := p.Header["Content-Type"]
	if len(c) > 0 {
		t, _, _ := mime.ParseMediaType(c[0])
		return t
	}
	return ""
}

func partIsAttachment(p *multipart.Part) (bool, map[string]string) {
	d := p.Header["Content-Disposition"]
	if len(d) > 0 {
		t, params, _ := mime.ParseMediaType(d[0])
		if t == "attachment" {
			return true, params
		}
	}
	return false, map[string]string{}
}
