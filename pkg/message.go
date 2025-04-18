package pkg

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
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

// Attachment is an image or other file attachment.
type Attachment struct {
	Content     []byte
	ContentType string
	FileName    string
}

// Message represents a matrix message.
type Message struct {
	Subject     string
	Body        string
	Preface     string
	Epilogue    string
	Attachments []Attachment
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

	body, attachments, err := parseBody(e)
	if err != nil {
		return nil, err
	}
	m.Body = string(body)
	m.Attachments = attachments

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

// parseBody will read the body from a mail.Message and returns the textual
// message content and any attachments, or an error if one is encountered.
func parseBody(m *mail.Message) ([]byte, []Attachment, error) {
	messageType, params, err := getMessageType(m)
	if err != nil {
		return nil, nil, err
	}

	if strings.HasPrefix(messageType, "multipart/") && params["boundary"] == "" {
		return nil, nil, fmt.Errorf("boundary parameter required for multipart message")
	}

	switch messageType {
	case "multipart/alternative":
		return readAlternativeParts(m, params["boundary"])
	case "multipart/mixed":
		return readMixedParts(m, params["boundary"])
	}

	b, err := io.ReadAll(m.Body)
	if err != nil {
		return nil, nil, err
	}

	// Send as text
	if strings.HasPrefix(messageType, "text/") {
		return b, nil, nil
	}

	// Send as attachment
	a := Attachment{
		ContentType: messageType,
		Content:     b,
	}
	return nil, []Attachment{a}, nil
}

// readAlternativeParts returns the content contained in the alternative parts.
// Only text/plain and text/html are recognized. In defiance of MIME (RFC2046),
// text/plain is preferred.
func readAlternativeParts(m *mail.Message, boundary string) ([]byte, []Attachment, error) {
	var txt []byte
	r := multipart.NewReader(m.Body, boundary)
	for {
		// NextPart() advances the reader, so if we need the content, we need to
		// read it before a subsequent call to NextPart.
		part, err := r.NextPart()
		if err != nil {
			break
		}
		mimeType, _, err := getPartContentType(part)
		if err != nil {
			return nil, nil, err
		}

		decoder := newPartDecoder(part)

		if mimeType == "text/html" {
			decoder = newHTMLSanitizer(decoder)
			b, err := io.ReadAll(decoder)
			if err != nil {
				return nil, nil, err
			}
			txt = b
		}

		if mimeType == "text/plain" {
			b, err := io.ReadAll(decoder)
			if err != nil {
				return nil, nil, err
			}
			txt = b
			break
		}
	}

	if txt == nil {
		return nil, nil, fmt.Errorf("unsupported alternative part")
	}

	return txt, nil, nil
}

// readMixedParts returns the textual content and any attachments contained in
// the mixed parts. Only text/plain and text/html are recognized for the
// textual portion. In defiance of MIME (RFC2046), text/plain is preferred. An
// error is returned if one is encountered.
func readMixedParts(m *mail.Message, boundary string) ([]byte, []Attachment, error) {
	txt := new(bytes.Buffer)
	attachments := []Attachment{}
	r := multipart.NewReader(m.Body, boundary)
	for {
		// NextPart() advances the reader, so if we need the content, we need to
		// read it before a subsequent call to NextPart.
		p, err := r.NextPart()
		if err != nil {
			break
		}

		disposition, dispositionParams, err := getPartDisposition(p)
		if err != nil {
			return nil, nil, err
		}
		mimeType, _, err := getPartContentType(p)
		if err != nil {
			return nil, nil, err
		}

		decoder := newPartDecoder(p)

		if disposition == "attachment" {
			b, err := io.ReadAll(decoder)
			if err != nil {
				return nil, nil, err
			}
			a := Attachment{
				ContentType: mimeType,
				Content:     b,
				FileName:    fmt.Sprint(dispositionParams["filename"]),
			}
			attachments = append(attachments, a)
			continue
		}

		if strings.HasPrefix(mimeType, "text/") {
			// Add newline between parts
			if txt.Len() > 0 {
				txt.Write([]byte("\n"))
			}
			if mimeType == "text/html" {
				decoder = newHTMLSanitizer(decoder)
			}
			decoder = newWhitespaceFixer(decoder)
			if _, err := io.Copy(txt, decoder); err != nil {
				return nil, nil, err
			}
		}
	}
	return txt.Bytes(), attachments, nil
}

// newHTMLSanitizer returns an io.Reader which emits the contents of r, with
// any HTML sanitized. The sanitized result is buffered at creation time, since
// r will be completely read immediately.
func newHTMLSanitizer(r io.Reader) io.Reader {
	policy := bluemonday.StrictPolicy()
	buf := policy.SanitizeReader(r)
	s := html.UnescapeString(buf.String())
	return newWhitespaceFixer(strings.NewReader(s))
}

// newWhitespaceFixer returns an io.Reader which emits the contents of r, but
// only allows a maximum of two newline characters in a row.
func newWhitespaceFixer(r io.Reader) io.Reader {
	return newRepeatRemover(r, '\n', 2)
}

// newPartDecoder returns an io.Reader which will provide the contents of p
// according to its content encoding.
func newPartDecoder(p *multipart.Part) io.Reader {
	switch getPartContentEncoding(p) {
	case "base64":
		return base64.NewDecoder(base64.StdEncoding, p)
	case "quoted-printable":
		return quotedprintable.NewReader(p)
	}
	// No decoding necessary
	return p
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

// getPartContentEncoding returns the content encoding of the part. If not set, the
// default according to RFC2045 6.1 (7bit) is returned.
func getPartContentEncoding(p *multipart.Part) string {
	if v := p.Header["Content-Transfer-Encoding"]; len(v) > 0 {
		return v[0]
	}
	return "7bit"
}

// getPartContentType returns the content type of the part. If not set, the
// default according to RFC2045 5.2 (text/plain) is returned.
func getPartContentType(p *multipart.Part) (contentType string, params map[string]string, err error) {
	if v := p.Header["Content-Type"]; len(v) > 0 {
		return mime.ParseMediaType(v[0])
	}
	return "text/plain", map[string]string{"charset": "us-ascii"}, nil
}

// getPartDisposition returns the content disposition of the part. If not set,
// the default according to RFC626 4.2 (inline) is returned.
func getPartDisposition(p *multipart.Part) (contentType string, params map[string]string, err error) {
	if v := p.Header["Content-Disposition"]; len(v) > 0 {
		return mime.ParseMediaType(v[0])
	}
	return "inline", map[string]string{}, nil
}
