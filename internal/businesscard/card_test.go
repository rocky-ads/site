package businesscard_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/rocky-ads/site/internal/businesscard"
)

func TestLoadCategoriesAlphabeticalIDs(t *testing.T) {
	cats, err := businesscard.LoadCategories("")
	if err != nil {
		t.Fatalf("LoadCategories: %v", err)
	}
	if len(cats) != 9 {
		t.Fatalf("expected 9 categories, got %d", len(cats))
	}
	if cats[0].Name != "Agricultural Equipment" || cats[0].ID != 1 {
		t.Fatalf("unexpected first category: %+v", cats[0])
	}
	if cats[3].Name != "Bicycles" || cats[3].ID != 4 {
		t.Fatalf("unexpected bicycles category: %+v", cats[3])
	}
}

func TestCategoryURL(t *testing.T) {
	got := businesscard.CategoryURL("rockyads.com", 4)
	want := "https://rockyads.com/c/4"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestRenderPrintCard(t *testing.T) {
	cats, err := businesscard.LoadCategories("")
	if err != nil {
		t.Fatalf("LoadCategories: %v", err)
	}
	cat, err := businesscard.FindCategory(cats, "Bicycles")
	if err != nil {
		t.Fatalf("FindCategory: %v", err)
	}

	img, err := businesscard.RenderPrintCard(businesscard.PrintCardOptions{
		Category:   cat,
		Host:       "rockyads.com",
		ImagesRoot: businesscard.ResolveRepoPath("static/images/category"),
	})
	if err != nil {
		t.Fatalf("RenderPrintCard: %v", err)
	}

	b := img.Bounds()
	if b.Dx() != 1125 || b.Dy() != 675 {
		t.Fatalf("expected 1125x675 with bleed, got %dx%d", b.Dx(), b.Dy())
	}

	var buf bytes.Buffer
	if err := businesscard.WritePNG(&buf, img); err != nil {
		t.Fatalf("WritePNG: %v", err)
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte{0x89, 'P', 'N', 'G'}) {
		t.Fatal("expected PNG output")
	}

	decoded, err := png.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("png decode: %v", err)
	}
	if decoded.Bounds().Dx() != b.Dx() {
		t.Fatal("decoded PNG size mismatch")
	}

	black := countColorNear(img, 0x00, 0x00, 0x00)
	if black < 500 {
		t.Fatalf("expected visible card content, got %d black pixels", black)
	}
}

func TestRenderPrintCardAllCategories(t *testing.T) {
	cats, err := businesscard.LoadCategories("")
	if err != nil {
		t.Fatalf("LoadCategories: %v", err)
	}
	for _, cat := range cats {
		name := cat.Name
		t.Run(name, func(t *testing.T) {
			_, err := businesscard.RenderPrintCard(businesscard.PrintCardOptions{
				Category:   cat,
				Host:       "rockyads.com",
				ImagesRoot: businesscard.ResolveRepoPath("static/images/category"),
			})
			if err != nil {
				t.Fatalf("RenderPrintCard: %v", err)
			}
		})
	}
}

func TestLongCategoryNameFits(t *testing.T) {
	cats, err := businesscard.LoadCategories("")
	if err != nil {
		t.Fatalf("LoadCategories: %v", err)
	}
	cat, err := businesscard.FindCategory(cats, "Agricultural Equipment Parts")
	if err != nil {
		t.Fatalf("FindCategory: %v", err)
	}
	_, err = businesscard.RenderPrintCard(businesscard.PrintCardOptions{
		Category:   cat,
		Host:       "rockyads.com",
		ImagesRoot: businesscard.ResolveRepoPath("static/images/category"),
	})
	if err != nil {
		t.Fatalf("RenderPrintCard: %v", err)
	}
}

func countColorNear(img *image.RGBA, r, g, b uint8) int {
	n := 0
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.RGBAAt(x, y)
			if colorNear(c, r, g, b) {
				n++
			}
		}
	}
	return n
}

func colorNear(c color.RGBA, r, g, b uint8) bool {
	const tol = 20
	return abs(int(c.R)-int(r)) <= tol &&
		abs(int(c.G)-int(g)) <= tol &&
		abs(int(c.B)-int(b)) <= tol &&
		c.A > 200
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func TestCategoryURLUsesHTTPS(t *testing.T) {
	url := businesscard.CategoryURL("rockyads.com", 1)
	if !strings.HasPrefix(url, "https://") {
		t.Fatalf("expected https URL, got %q", url)
	}
}
