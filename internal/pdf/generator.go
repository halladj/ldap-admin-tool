package pdf

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"

	"github.com/jung-kurt/gofpdf"
	"github.com/halladj/ldap-admin-tool/internal/config"
	"github.com/halladj/ldap-admin-tool/internal/types"
)

//go:embed assets/logo.png
var logoPNG []byte

func Generate(cfg *config.Config, user types.User) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")

	// Register embedded logo
	pdf.RegisterImageOptionsReader("logo", gofpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(logoPNG))

	pdf.AddPage()

	// Logo centered at top
	imgW := 50.0
	pageW, _ := pdf.GetPageSize()
	pdf.ImageOptions("logo", (pageW-imgW)/2, 10, imgW, 0, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	pdf.Ln(32)

	// Lab name
	pdf.SetFont("Helvetica", "B", 18)
	pdf.CellFormat(0, 10, "MISC Laboratory", "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "", 13)
	pdf.CellFormat(0, 8, "LDAP Account Credentials", "", 1, "C", false, 0, "")
	pdf.Ln(8)

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
		{"Services Portal", "services.misc-lab.org"},
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
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	if err := pdf.OutputFileAndClose(tmpPath); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read generated PDF: %w", err)
	}

	return data, nil
}
