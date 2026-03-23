package qrcode

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg" // Support JPEG logos
	"image/png"
	"os"
	"strings"

	"github.com/nfnt/resize" // I need to install this for resizing the logo
	qrcode "github.com/skip2/go-qrcode"
)

// GenerateSVG generates a QR code in SVG format (Infinite quality, tiny file size)
func GenerateSVG(content string, size int) (string, error) {
	if content == "" {
		return "", fmt.Errorf("content cannot be empty")
	}
	qr, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return "", err
	}

	// Manual SVG construction from the bit matrix
	matrix := qr.Bitmap()
	moduleCount := len(matrix)
	
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" shape-rendering="crispEdges">`, moduleCount, moduleCount))
	sb.WriteString(`<path fill="#ffffff" d="M0 0h%v v%vh-%v z"/>`)
	sb.WriteString(`<path fill="#000000" d="`)
	
	for y, row := range matrix {
		var spanStart int
		inSpan := false
		
		for x, active := range row {
			if active && !inSpan {
				spanStart = x
				inSpan = true
			} else if !active && inSpan {
				// End of black span, write as one horizontal line
				sb.WriteString(fmt.Sprintf("M%d %dh%dv1h-%dz", spanStart, y, x-spanStart, x-spanStart))
				inSpan = false
			}
		}
		// Handle span at the end of the row
		if inSpan {
			width := len(row) - spanStart
			sb.WriteString(fmt.Sprintf("M%d %dh%dv1h-%dz", spanStart, y, width, width))
		}
	}
	sb.WriteString(`"/></svg>`)

	// Final SVG string with moduleCount applied
	return fmt.Sprintf(sb.String(), moduleCount, moduleCount, moduleCount), nil
}

// GenerateBrandedBase64 generates a QR code with a company logo in the center
func GenerateBrandedBase64(content string, logoPath string, size int) (string, error) {
	if content == "" {
		return "", fmt.Errorf("content cannot be empty")
	}
	if size <= 0 {
		size = 512 // Higher resolution for branded QR
	}

	// 1. Generate the base QR code
	qr, err := qrcode.New(content, qrcode.High) // High recovery level is mandatory for branded QRs
	if err != nil {
		return "", err
	}
	qrImage := qr.Image(size)

	// 2. Load the logo
	logoFile, err := os.Open(logoPath)
	if err != nil {
		return "", fmt.Errorf("failed to open logo: %w", err)
	}
	defer logoFile.Close()

	logoImg, _, err := image.Decode(logoFile)
	if err != nil {
		return "", fmt.Errorf("failed to decode logo: %w", err)
	}

	// 3. Resize logo to be ~20% of the QR code size
	logoSize := uint(size / 5)
	logoImg = resize.Resize(logoSize, logoSize, logoImg, resize.Lanczos3)

	// 4. Create final canvas
	brandedQR := image.NewRGBA(qrImage.Bounds())
	draw.Draw(brandedQR, qrImage.Bounds(), qrImage, image.Point{}, draw.Src)

	// 5. Calculate center position
	offset := image.Pt((size-int(logoSize))/2, (size-int(logoSize))/2)

	// 6. Draw white buffer around logo for better scannability
	bufferSize := int(logoSize) + 10
	bufferOffset := image.Pt((size-bufferSize)/2, (size-bufferSize)/2)
	whiteBuffer := image.NewUniform(image.White)
	draw.Draw(brandedQR, image.Rect(bufferOffset.X, bufferOffset.Y, bufferOffset.X+bufferSize, bufferOffset.Y+bufferSize), whiteBuffer, image.Point{}, draw.Src)

	// 7. Draw the actual logo
	draw.Draw(brandedQR, image.Rect(offset.X, offset.Y, offset.X+int(logoSize), offset.Y+int(logoSize)), logoImg, image.Point{}, draw.Over)

	// 8. Encode to Base64
	var buf bytes.Buffer
	err = png.Encode(&buf, brandedQR)
	if err != nil {
		return "", err
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// GenerateBase64 generates a plain QR code
func GenerateBase64(content string, size int) (string, error) {
	pngData, err := qrcode.Encode(content, qrcode.Medium, size)
	if err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngData), nil
}

// GenerateUPILink creates a standard UPI payment URI
func GenerateUPILink(upiID, name, amount, note string) string {
	return fmt.Sprintf("upi://pay?pa=%s&pn=%s&am=%s&tn=%s", upiID, name, amount, note)
}
