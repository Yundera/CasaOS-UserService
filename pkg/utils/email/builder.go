package email

import (
	"bytes"
	"embed"
	"encoding/base64"
	"fmt"
	"time"
)

//go:embed assets/yundera-logo.png
var logoFS embed.FS

// InlineAttachment represents an inline image for emails
type InlineAttachment struct {
	Content   []byte // Raw binary content
	MimeType  string // e.g., "image/png"
	ContentID string // e.g., "yundera_logo" (referenced as cid:yundera_logo in HTML)
	Filename  string // e.g., "logo.png"
}

// EmailContent holds the email data for building
type EmailContent struct {
	Subject     string
	TextBody    string
	HTMLBody    string
	Attachments []InlineAttachment
}

// YunderaLogo returns the Yundera logo as an InlineAttachment
func YunderaLogo() InlineAttachment {
	content, _ := logoFS.ReadFile("assets/yundera-logo.png")
	return InlineAttachment{
		Content:   content,
		MimeType:  "image/png",
		ContentID: "yundera_logo",
		Filename:  "yundera-logo.png",
	}
}

// YunderaLogoContentID is the Content-ID reference for the Yundera logo
// Use this in HTML as: <img src="cid:yundera_logo" />
const YunderaLogoContentID = "cid:yundera_logo"

// BuildMIMEEmail constructs a proper multipart MIME email with inline attachments
// Returns the complete email message as bytes ready for SMTP
func BuildMIMEEmail(from, to string, content EmailContent) ([]byte, error) {
	var buf bytes.Buffer

	// Generate unique boundaries
	relatedBoundary := generateBoundary("related")
	alternativeBoundary := generateBoundary("alternative")

	// Write basic headers
	buf.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", to))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", content.Subject))
	buf.WriteString("MIME-Version: 1.0\r\n")

	if len(content.Attachments) > 0 {
		// multipart/related wraps multipart/alternative + inline attachments
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/related; boundary=\"%s\"\r\n", relatedBoundary))
		buf.WriteString("\r\n")

		// Start related section with alternative content
		buf.WriteString(fmt.Sprintf("--%s\r\n", relatedBoundary))
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", alternativeBoundary))
		buf.WriteString("\r\n")

		// Plain text part
		buf.WriteString(fmt.Sprintf("--%s\r\n", alternativeBoundary))
		buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		buf.WriteString("Content-Transfer-Encoding: 7bit\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(content.TextBody)
		buf.WriteString("\r\n")

		// HTML part
		buf.WriteString(fmt.Sprintf("--%s\r\n", alternativeBoundary))
		buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		buf.WriteString("Content-Transfer-Encoding: 7bit\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(content.HTMLBody)
		buf.WriteString("\r\n")

		// Close alternative boundary
		buf.WriteString(fmt.Sprintf("--%s--\r\n", alternativeBoundary))

		// Add inline attachments
		for _, att := range content.Attachments {
			buf.WriteString(fmt.Sprintf("--%s\r\n", relatedBoundary))
			buf.WriteString(fmt.Sprintf("Content-Type: %s; name=\"%s\"\r\n", att.MimeType, att.Filename))
			buf.WriteString("Content-Transfer-Encoding: base64\r\n")
			buf.WriteString(fmt.Sprintf("Content-ID: <%s>\r\n", att.ContentID))
			buf.WriteString(fmt.Sprintf("Content-Disposition: inline; filename=\"%s\"\r\n", att.Filename))
			buf.WriteString("\r\n")

			// Write base64 content with line wrapping (76 chars per line per RFC)
			encoded := base64.StdEncoding.EncodeToString(att.Content)
			for i := 0; i < len(encoded); i += 76 {
				end := i + 76
				if end > len(encoded) {
					end = len(encoded)
				}
				buf.WriteString(encoded[i:end])
				buf.WriteString("\r\n")
			}
		}

		// Close related boundary
		buf.WriteString(fmt.Sprintf("--%s--\r\n", relatedBoundary))
	} else {
		// No attachments - simple multipart/alternative
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", alternativeBoundary))
		buf.WriteString("\r\n")

		// Plain text part
		buf.WriteString(fmt.Sprintf("--%s\r\n", alternativeBoundary))
		buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		buf.WriteString("Content-Transfer-Encoding: 7bit\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(content.TextBody)
		buf.WriteString("\r\n")

		// HTML part
		buf.WriteString(fmt.Sprintf("--%s\r\n", alternativeBoundary))
		buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		buf.WriteString("Content-Transfer-Encoding: 7bit\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(content.HTMLBody)
		buf.WriteString("\r\n")

		// Close alternative boundary
		buf.WriteString(fmt.Sprintf("--%s--\r\n", alternativeBoundary))
	}

	return buf.Bytes(), nil
}

// generateBoundary creates a unique boundary string for MIME
func generateBoundary(prefix string) string {
	return fmt.Sprintf("=_%s_%d_boundary=", prefix, time.Now().UnixNano())
}
