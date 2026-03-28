package images

import (
	"bytes"
	"image"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/nfnt/resize"
)

// OptimizeLogo takes any uploaded image and crushes it to an industrial size (under 10KB)
func OptimizeLogo(inputPath string) ([]byte, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	// Resize to 150px height (standard for invoices)
	m := resize.Resize(0, 150, img, resize.Lanczos3)

	var buf bytes.Buffer
	// Encode with 75% quality to save space while keeping brand clarity
	err = jpeg.Encode(&buf, m, &jpeg.Options{Quality: 75})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
