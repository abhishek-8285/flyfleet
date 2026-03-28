package vectorengine

import (
	"bytes"
	"text/template"

	"github.com/go-pdf/fpdf"
	"github.com/skip2/go-qrcode"
)

type LayoutElement struct {
	Type       string // "text", "line", "rect", "qrcode", "image"
	X, Y, W, H float64
	Text       string
	Font       string
	Style      string
	Size       float64
	Color      []int  // [R, G, B]
	FillColor  []int  // [R, G, B]
	Align      string // "L", "C", "R"
}

type Layout struct {
	Elements []LayoutElement
}

type Engine struct {
	pdf *fpdf.Fpdf
}

func NewEngine() *Engine {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetCompression(true) // Crucial for minimum size
	return &Engine{pdf: pdf}
}

func (e *Engine) Generate(layout Layout, data interface{}) ([]byte, error) {
	e.pdf.AddPage()

	for _, el := range layout.Elements {
		switch el.Type {
		case "rect":
			if len(el.FillColor) == 3 {
				e.pdf.SetFillColor(el.FillColor[0], el.FillColor[1], el.FillColor[2])
				e.pdf.Rect(el.X, el.Y, el.W, el.H, "F")
			} else {
				e.pdf.SetDrawColor(200, 200, 200)
				e.pdf.Rect(el.X, el.Y, el.W, el.H, "D")
			}
		case "text":
			e.drawText(el, data)
		case "line":
			e.pdf.SetDrawColor(180, 180, 180)
			e.pdf.Line(el.X, el.Y, el.X+el.W, el.Y+el.H)
		case "qrcode":
			e.drawQRCode(el, data)
		case "image":
			// Text field stores the raw byte path or buffer
			e.pdf.RegisterImageOptions(el.Text, fpdf.ImageOptions{ImageType: "JPG", ReadDpi: true})
			e.pdf.ImageOptions(el.Text, el.X, el.Y, el.W, el.H, false, fpdf.ImageOptions{ImageType: "JPG", ReadDpi: true}, 0, "")
		}
	}

	var buf bytes.Buffer
	err := e.pdf.Output(&buf)
	return buf.Bytes(), err
}

func (e *Engine) drawText(el LayoutElement, data interface{}) {
	tmpl, err := template.New("t").Parse(el.Text)
	if err != nil {
		return
	}
	var tBuf bytes.Buffer
	tmpl.Execute(&tBuf, data)
	txt := tBuf.String()

	if len(el.Color) == 3 {
		e.pdf.SetTextColor(el.Color[0], el.Color[1], el.Color[2])
	} else {
		e.pdf.SetTextColor(30, 30, 30)
	}

	font := "Arial"
	if el.Font != "" {
		font = el.Font
	}
	e.pdf.SetFont(font, el.Style, el.Size)
	e.pdf.SetXY(el.X, el.Y)

	// Handle alignment correctly
	align := "L"
	if el.Align != "" {
		align = el.Align
	}

	e.pdf.CellFormat(el.W, el.H, txt, "", 0, align, false, 0, "")
}

func (e *Engine) drawQRCode(el LayoutElement, data interface{}) {
	tmpl, _ := template.New("q").Parse(el.Text)
	var tBuf bytes.Buffer
	tmpl.Execute(&tBuf, data)

	qr, _ := qrcode.New(tBuf.String(), qrcode.Medium)
	matrix := qr.Bitmap()
	modSize := el.W / float64(len(matrix))
	e.pdf.SetFillColor(0, 0, 0)
	for y, row := range matrix {
		for x, active := range row {
			if active {
				e.pdf.Rect(el.X+(float64(x)*modSize), el.Y+(float64(y)*modSize), modSize, modSize, "F")
			}
		}
	}
}
