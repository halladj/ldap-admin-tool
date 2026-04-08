package pdf

import (
	"fmt"
	"os"

	"github.com/jung-kurt/gofpdf"
	"github.com/halladj/ldap-admin-tool/internal/config"
	"github.com/halladj/ldap-admin-tool/internal/types"
)

func Generate(cfg *config.Config, user types.User) (string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Header
	pdf.SetFont("Helvetica", "B", 20)
	pdf.CellFormat(0, 15, "your-domain.org", "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "", 14)
	pdf.CellFormat(0, 10, "LDAP Account Credentials", "", 1, "C", false, 0, "")
	pdf.Ln(10)

	// Account Details header
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, 10, "Account Details", "", 1, "L", false, 0, "")
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(5)

	// Fields
	fields := []struct {
		Label string
		Value string
	}{
		{"Full Name", fmt.Sprintf("%s %s", user.FirstName, user.LastName)},
		{"Username", user.UID},
		{"Email", user.Email},
		{"Password", user.Password},
		{"LDAP Server", cfg.LDAPServer},
		{"Base DN", cfg.BaseDN},
	}

	for _, f := range fields {
		pdf.SetFont("Helvetica", "B", 11)
		pdf.CellFormat(45, 8, f.Label+":", "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 11)
		pdf.CellFormat(0, 8, f.Value, "", 1, "L", false, 0, "")
	}

	// Groups
	if len(user.Groups) > 0 {
		pdf.Ln(5)
		pdf.SetFont("Helvetica", "B", 12)
		pdf.CellFormat(0, 10, "Group Memberships", "", 1, "L", false, 0, "")
		pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
		pdf.Ln(5)
		pdf.SetFont("Helvetica", "", 11)
		for _, g := range user.Groups {
			pdf.CellFormat(0, 8, "  - "+g, "", 1, "L", false, 0, "")
		}
	}

	// Footer
	pdf.Ln(15)
	pdf.SetFont("Helvetica", "I", 9)
	pdf.CellFormat(0, 8, "This document is confidential. Please change your password after first login.", "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 8, "This is an automatically generated document.", "", 1, "C", false, 0, "")

	tmpFile, err := os.CreateTemp("", fmt.Sprintf("ldap_%s_*.pdf", user.UID))
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpFile.Close()

	if err := pdf.OutputFileAndClose(tmpFile.Name()); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	return tmpFile.Name(), nil
}
