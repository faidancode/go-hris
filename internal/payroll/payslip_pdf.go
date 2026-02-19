package payroll

import (
	"bytes"
	"fmt"
	"strings"
)

func buildSimplePayslipPDF(lines []string) ([]byte, error) {
	if len(lines) == 0 {
		lines = []string{"Payslip"}
	}

	var content strings.Builder
	content.WriteString("BT\n/F1 12 Tf\n50 800 Td\n")
	for i, line := range lines {
		escaped := pdfEscape(line)
		if i == 0 {
			content.WriteString(fmt.Sprintf("(%s) Tj\n", escaped))
			continue
		}
		content.WriteString(fmt.Sprintf("T* (%s) Tj\n", escaped))
	}
	content.WriteString("ET")

	stream := content.String()
	objects := []string{
		"1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n",
		"2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n",
		"3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>\nendobj\n",
		"4 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n",
		fmt.Sprintf("5 0 obj\n<< /Length %d >>\nstream\n%s\nendstream\nendobj\n", len(stream), stream),
	}

	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n")
	offsets := make([]int, 0, len(objects)+1)
	offsets = append(offsets, 0)

	for _, obj := range objects {
		offsets = append(offsets, out.Len())
		out.WriteString(obj)
	}

	xrefStart := out.Len()
	out.WriteString(fmt.Sprintf("xref\n0 %d\n", len(offsets)))
	out.WriteString("0000000000 65535 f \n")
	for i := 1; i < len(offsets); i++ {
		out.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}
	out.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF", len(offsets), xrefStart))

	return out.Bytes(), nil
}

func pdfEscape(v string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "(", "\\(", ")", "\\)")
	return replacer.Replace(v)
}
