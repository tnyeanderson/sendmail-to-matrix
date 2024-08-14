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

	"github.com/Masterminds/sprig"
	"github.com/microcosm-cc/bluemonday"
)

// DefaultMessageTemplate is the default template used to render messages.
const DefaultMessageTemplate = `{{.Preface}}
Subject: {{.Subject}}

{{.Body}}
{{.Epilogue}}`

// Message represents a matrix message.
type Message struct {
	Subject  string
	Body     string
	Preface  string
	Epilogue string
}

// NewMessage reads an email from an io.Reader (usually stdin) and returns a
// Message with the data.
func NewMessage(r io.Reader) (*Message, error) {
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
// template.
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

	return out.Bytes(), nil
}

// FilterMessage returns true if the message should be forwarded to matrix and
// false if the message should be skipped/ignored.
func FilterMessage(message string, skips []*regexp.Regexp) bool {
	for _, r := range skips {
		if r.MatchString(message) {
			return false
		}
	}
	return true
}

func parseBody(m *mail.Message) (string, error) {
	messageType, params := getMessageType(m)
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

func removeHtmlTags(input []byte) []byte {
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
		bytes = removeHtmlTags(bytes)
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
			// Add newline between parts
			if out.Len() > 0 {
				out.Write([]byte("\n"))
			}
			if partType == "text/html" {
				buf := bytes.NewBuffer([]byte{})
				io.Copy(buf, p)
				b := removeHtmlTags(buf.Bytes())
				out.Write(b)
				continue
			}
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
