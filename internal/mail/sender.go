package mail

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"os/exec"
)

// WelcomeEmail contains the information needed to send a welcome email with credentials.
type WelcomeEmail struct {
	From      string
	To        string
	FirstName string
	LastName  string
	UID       string
	PDF       []byte
}

// SendWelcome sends a welcome email with LDAP credentials and PDF attachment.
func SendWelcome(e WelcomeEmail) error {
	body := fmt.Sprintf(`Hello %s %s,

Your LDAP account on your-domain.org has been created.

Your credentials are attached as a PDF document. Please keep them safe.

Username: %s

Please change your password after your first login.

--
your-domain.org Admin Team

This is an automatically generated email, please do not reply.
`, e.FirstName, e.LastName, e.UID)

	var msg bytes.Buffer
	mw := multipart.NewWriter(&msg)

	// Email headers
	fmt.Fprintf(&msg, "From: %s\r\n", e.From)
	fmt.Fprintf(&msg, "To: %s\r\n", e.To)
	fmt.Fprintf(&msg, "Subject: Your your-domain.org LDAP Account Credentials\r\n")
	fmt.Fprintf(&msg, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&msg, "Content-Type: multipart/mixed; boundary=%s\r\n\r\n", mw.Boundary())

	// Text part
	textHeader := make(textproto.MIMEHeader)
	textHeader.Set("Content-Type", "text/plain; charset=utf-8")
	part, err := mw.CreatePart(textHeader)
	if err != nil {
		return fmt.Errorf("failed to create text part: %w", err)
	}
	fmt.Fprintf(part, "%s\r\n", body)

	// PDF attachment
	pdfHeader := make(textproto.MIMEHeader)
	pdfHeader.Set("Content-Type", `application/pdf; name="ldap_credentials.pdf"`)
	pdfHeader.Set("Content-Transfer-Encoding", "base64")
	pdfHeader.Set("Content-Disposition", `attachment; filename="ldap_credentials.pdf"`)
	pdfPart, err := mw.CreatePart(pdfHeader)
	if err != nil {
		return fmt.Errorf("failed to create PDF part: %w", err)
	}
	enc := base64.NewEncoder(base64.StdEncoding, pdfPart)
	if _, err := enc.Write(e.PDF); err != nil {
		return fmt.Errorf("failed to encode PDF: %w", err)
	}
	enc.Close()

	mw.Close()

	cmd := exec.Command("/usr/sbin/sendmail", "-t", "-oi")
	cmd.Stdin = &msg

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("sendmail failed: %w: %s", err, string(output))
	}

	return nil
}
