package main
import (
	"image"
	_ "image/jpeg"
	"image/jpeg"
	"os"
	"github.com/nfnt/resize"
)
func main() {
	file, _ := os.Open("logo1.jpg")
	img, _, _ := image.Decode(file)
	m := resize.Resize(200, 0, img, resize.Lanczos3)
	out, _ := os.Create("logo_tiny.jpg")
	jpeg.Encode(out, m, &jpeg.Options{Quality: 80})
}
