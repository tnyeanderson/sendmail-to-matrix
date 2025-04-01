package pkg

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"regexp"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/microcosm-cc/bluemonday"
)

// DefaultMessageTemplate is the default template used to render messages.
const DefaultMessageTemplate = `
{{- if .Preface }}{{ println .Preface }}{{ end -}}
{{- if .Subject }}{{ printf "Subject: %s\n" .Subject }}{{ end -}}
{{- if .Body }}{{ println .Body }}{{ end -}}
{{- if .Epilogue }}{{ println .Epilogue }}{{ end -}}
`

// Message represents a matrix message.
type Message struct {
	Subject  string
	Body     string
	Preface  string
	Epilogue string
}

// NewMessageFromEmail reads an email from an io.Reader (usually stdin) and returns a
// Message with the data.
func NewMessageFromEmail(r io.Reader) (*Message, error) {
	m := &Message{}
	e, err := mail.ReadMessage(r)
	if err != nil {
		return nil, err
	}
	m.Subject = e.Header.Get("Subject")

	body, err := parseBody(e)
	if err != nil {
		return nil, err
	}
	m.Body = body

	return m, nil
}

// Render generates the message text to be sent based on a Message and a go
// template. Leading and trailing newlines are trimmed.
func (m *Message) Render(templateText []byte) ([]byte, error) {
	name := "stm"

	// Create template
	t, err := template.New(name).Funcs(sprig.FuncMap()).Parse(string(templateText))
	if err != nil {
		return nil, err
	}

	// Execute template
	out := bytes.Buffer{}
	if err := t.ExecuteTemplate(&out, name, *m); err != nil {
		return nil, err
	}

	return bytes.TrimSpace(out.Bytes()), nil
}

func parseBody(m *mail.Message) (string, error) {
	messageType, params, err := getMessageType(m)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(messageType, "multipart/") {
		boundary := params["boundary"]
		if boundary != "" {
			content, err := parseMultipart(m, messageType, boundary)
			if err == nil {
				return string(content), nil
			}
		}
	}
	b, err := io.ReadAll(m.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func removeHTMLTags(input []byte) []byte {
	policy := bluemonday.StrictPolicy()
	s := policy.Sanitize(string(input))
	s = html.UnescapeString(s)
	s = fixWhitespace(s)
	return []byte(s)
}

func fixWhitespace(message string) string {
	re := regexp.MustCompile("\n\n+")
	return re.ReplaceAllLiteralString(message, "\n\n")
}

func parseMultipart(m *mail.Message, messageType, boundary string) ([]byte, error) {
	mr := multipart.NewReader(m.Body, boundary)
	if messageType == "multipart/alternative" {
		return readAlternativeParts(mr)
	}
	if messageType == "multipart/mixed" {
		return readMixedParts(mr)
	}
	return nil, fmt.Errorf("not a recognized multipart message")
}

// readAlternativeParts returns the content contained in the alternative parts.
// Only text/plain and text/html are recognized. In defiance of MIME (RFC2046),
// text/plain is preferred.
func readAlternativeParts(r *multipart.Reader) ([]byte, error) {
	parts := map[string][]byte{}
	var out []byte
	for {
		// NextPart() advances the reader, so if we need the content, we need to
		// read it before a subsequent call to NextPart.
		part, err := r.NextPart()
		if err != nil {
			break
		}
		mimeType, _, err := getPartType(part)
		if err != nil {
			return nil, err
		}

		if mimeType == "text/plain" {
			b, err := io.ReadAll(part)
			if err != nil {
				return nil, err
			}
			out = b
			continue
		}

		// If we already have a text/plain, ignore other MIME types.
		if _, ok := parts["text/plain"]; ok {
			continue
		}

		if mimeType == "text/html" {
			b, err := io.ReadAll(part)
			if err != nil {
				return nil, err
			}
			out = removeHTMLTags(b)
		}
	}

	if out == nil {
		return nil, fmt.Errorf("unsupported alternative part")
	}

	return out, nil
}

func readMixedParts(r *multipart.Reader) ([]byte, error) {
	out := new(bytes.Buffer)
	for {
		p, err := r.NextPart()
		if err != nil {
			break
		}
		mimeType, _, err := getPartType(p)
		if err != nil {
			return nil, err
		}
		// Only text parts are recognized.
		// TODO: Handle attachments
		if strings.HasPrefix(mimeType, "text/") {
			// Add newline between parts
			if out.Len() > 0 {
				out.Write([]byte("\n"))
			}
			if mimeType == "text/html" {
				b, err := io.ReadAll(p)
				if err != nil {
					return nil, err
				}
				out.Write(removeHTMLTags(b))
				continue
			}
			if _, err := io.Copy(out, p); err != nil {
				return nil, err
			}
		}
	}
	return out.Bytes(), nil
}

// getMessageType returns the top-level media type and parameters. If not set,
// the default according to RFC2045 5.2 (text/plain) is returned.
func getMessageType(m *mail.Message) (contentType string, params map[string]string, err error) {
	c := m.Header["Content-Type"]
	if len(c) > 0 {
		return mime.ParseMediaType(c[0])
	}
	return "text/plain", map[string]string{"charset": "us-ascii"}, nil
}

// getPartType returns the content type of the part. If not set, the default
// according to RFC2045 5.2 (text/plain) is returned.
func getPartType(p *multipart.Part) (contentType string, params map[string]string, err error) {
	if v := p.Header["Content-Disposition"]; len(v) > 0 {
		contentType, params, err = mime.ParseMediaType(v[0])
		if contentType == "attachment" {
			return
		}
	}
	if v := p.Header["Content-Type"]; len(v) > 0 {
		return mime.ParseMediaType(v[0])
	}
	return "text/plain", map[string]string{"charset": "us-ascii"}, nil
}
