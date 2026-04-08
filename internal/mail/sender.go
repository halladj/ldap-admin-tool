package mail

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"os"
	"os/exec"
	"path/filepath"

)

func SendWelcome(senderEmail, to, firstName, lastName, uid, pdfPath string) error {
	body := fmt.Sprintf(`Hello %s %s,

Your LDAP account on your-domain.org has been created.

Your credentials are attached as a PDF document. Please keep them safe.

Username: %s

Please change your password after your first login.

--
your-domain.org Admin Team

This is an automatically generated email, please do not reply.
`, firstName, lastName, uid)

	pdfData, err := os.ReadFile(pdfPath)
	if err != nil {
		return fmt.Errorf("failed to read PDF: %w", err)
	}

	var msg bytes.Buffer
	writer := multipart.NewWriter(&msg)

	// Headers
	boundary := writer.Boundary()
	msg.Reset()
	fmt.Fprintf(&msg, "From: %s\r\n", senderEmail)
	fmt.Fprintf(&msg, "To: %s\r\n", to)
	fmt.Fprintf(&msg, "Subject: Your your-domain.org LDAP Account Credentials\r\n")
	fmt.Fprintf(&msg, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&msg, "Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary)
	fmt.Fprintf(&msg, "\r\n")

	// Text body
	fmt.Fprintf(&msg, "--%s\r\n", boundary)
	fmt.Fprintf(&msg, "Content-Type: text/plain; charset=utf-8\r\n")
	fmt.Fprintf(&msg, "\r\n")
	fmt.Fprintf(&msg, "%s\r\n", body)

	// PDF attachment
	fmt.Fprintf(&msg, "--%s\r\n", boundary)

	attachHeader := make(textproto.MIMEHeader)
	attachHeader.Set("Content-Type", "application/pdf")
	attachHeader.Set("Content-Transfer-Encoding", "base64")
	attachHeader.Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base("ldap_credentials_"+uid+".pdf")))

	for key, values := range attachHeader {
		for _, v := range values {
			fmt.Fprintf(&msg, "%s: %s\r\n", key, v)
		}
	}
	fmt.Fprintf(&msg, "\r\n")

	encoded := base64.StdEncoding.EncodeToString(pdfData)
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		fmt.Fprintf(&msg, "%s\r\n", encoded[i:end])
	}

	fmt.Fprintf(&msg, "--%s--\r\n", boundary)

	cmd := exec.Command("/usr/sbin/sendmail", "-t", "-oi")
	cmd.Stdin = &msg

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sendmail failed: %w: %s", err, string(output))
	}

	return nil
}
