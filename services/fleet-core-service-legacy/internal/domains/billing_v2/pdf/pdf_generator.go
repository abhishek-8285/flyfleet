package pdf

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// PDFOptions represents PDF generation options
type PDFOptions struct {
	Landscape           bool    `json:"landscape"`
	PrintBackground     bool    `json:"printBackground"`
	Scale               float64 `json:"scale"`
	PaperWidth          float64 `json:"paperWidth"`  // inches
	PaperHeight         float64 `json:"paperHeight"` // inches
	MarginTop           float64 `json:"marginTop"`   // inches
	MarginBottom        float64 `json:"marginBottom"`
	MarginLeft          float64 `json:"marginLeft"`
	MarginRight         float64 `json:"marginRight"`
	PageRanges          string  `json:"pageRanges"`
	IgnoreInvalidPageRanges bool `json:"ignoreInvalidPageRanges"`
	HeaderTemplate      string  `json:"headerTemplate"`
	FooterTemplate      string  `json:"footerTemplate"`
	PreferCSSPageSize   bool    `json:"preferCSSPageSize"`
}

// DefaultA4Portrait returns default A4 portrait options
func DefaultA4Portrait() PDFOptions {
	return PDFOptions{
		Landscape:       false,
		PrintBackground: true,
		Scale:           1.0,
		PaperWidth:      8.27,
		PaperHeight:     11.69,
		MarginTop:       0.4,
		MarginBottom:    0.4,
		MarginLeft:      0.4,
		MarginRight:     0.4,
	}
}

// PDFGenerator generates PDFs using chromedp
type PDFGenerator struct {
	timeout time.Duration
}

// NewPDFGenerator creates a new PDF generator
func NewPDFGenerator(timeout time.Duration) *PDFGenerator {
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	return &PDFGenerator{timeout: timeout}
}

// GenerateFromHTML generates a PDF from HTML content
func (g *PDFGenerator) GenerateFromHTML(ctx context.Context, html string, options PDFOptions) ([]byte, error) {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, g.timeout)
	defer cancel()

	var buf []byte
	err := chromedp.Run(ctx,
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}
			return page.SetDocumentContent(frameTree.Frame.ID, html).Do(ctx)
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			printParams := page.PrintToPDF().
				WithLandscape(options.Landscape).
				WithPrintBackground(options.PrintBackground).
				WithScale(options.Scale).
				WithPaperWidth(options.PaperWidth).
				WithPaperHeight(options.PaperHeight).
				WithMarginTop(options.MarginTop).
				WithMarginBottom(options.MarginBottom).
				WithMarginLeft(options.MarginLeft).
				WithMarginRight(options.MarginRight).
				WithHeaderTemplate(options.HeaderTemplate).
				WithFooterTemplate(options.FooterTemplate).
				WithPreferCSSPageSize(options.PreferCSSPageSize)

			if options.HeaderTemplate != "" || options.FooterTemplate != "" {
				printParams = printParams.WithDisplayHeaderFooter(true)
			}

			data, _, err := printParams.Do(ctx)
			if err != nil {
				return err
			}
			buf = data
			return nil
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("chromedp failed: %w", err)
	}

	return buf, nil
}

// GenerateFromURL generates a PDF from a URL
func (g *PDFGenerator) GenerateFromURL(ctx context.Context, url string, options PDFOptions) ([]byte, error) {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, g.timeout)
	defer cancel()

	var buf []byte
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			printParams := page.PrintToPDF().
				WithLandscape(options.Landscape).
				WithPrintBackground(options.PrintBackground).
				WithScale(options.Scale).
				WithPaperWidth(options.PaperWidth).
				WithPaperHeight(options.PaperHeight).
				WithMarginTop(options.MarginTop).
				WithMarginBottom(options.MarginBottom).
				WithMarginLeft(options.MarginLeft).
				WithMarginRight(options.MarginRight).
				WithHeaderTemplate(options.HeaderTemplate).
				WithFooterTemplate(options.FooterTemplate).
				WithPreferCSSPageSize(options.PreferCSSPageSize)

			if options.HeaderTemplate != "" || options.FooterTemplate != "" {
				printParams = printParams.WithDisplayHeaderFooter(true)
			}

			data, _, err := printParams.Do(ctx)
			if err != nil {
				return err
			}
			buf = data
			return nil
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("chromedp failed: %w", err)
	}

	return buf, nil
}

// GenerateFromTemplate generates a PDF using a Go template and data
func (g *PDFGenerator) GenerateFromTemplate(ctx context.Context, templateHTML string, data interface{}, options PDFOptions) ([]byte, error) {
	tmpl, err := template.New("invoice").Parse(templateHTML)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return g.GenerateFromHTML(ctx, buf.String(), options)
}

// GenerateInvoicePDF generates an invoice PDF using a template
func (g *PDFGenerator) GenerateInvoicePDF(ctx context.Context, invoiceData map[string]interface{}, templateHTML string) ([]byte, error) {
	return g.GenerateFromTemplate(ctx, templateHTML, invoiceData, DefaultA4Portrait())
}

// EncodeToBase64 encodes PDF data to base64
func EncodeToBase64(pdfData []byte) string {
	return base64.StdEncoding.EncodeToString(pdfData)
}

// SaveToFile saves PDF data to a file
func SaveToFile(pdfData []byte, filename string) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return os.WriteFile(filename, pdfData, 0644)
}
