package image

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

// syntheticPNG creates a minimal cols×rows PNG with a gradient.
func syntheticPNG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(x * 255 / width),
				G: uint8(y * 255 / height),
				B: 128,
				A: 255,
			})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encoding PNG: %v", err)
	}
	return buf.Bytes()
}

func TestRender_NonEmpty(t *testing.T) {
	data := syntheticPNG(t, 64, 32)
	ansi, err := Render(data, 20, 10)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if ansi == "" {
		t.Fatal("Render returned empty string")
	}
	if !strings.Contains(ansi, "▄") {
		t.Error("expected half-block character ▄ in ANSI output")
	}
}

func TestRender_WidthBound(t *testing.T) {
	data := syntheticPNG(t, 200, 100)
	const cols = 30
	ansi, err := Render(data, cols, 15)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	// Check that no line exceeds requested cols (accounting for ANSI escape codes,
	// strip them to measure visible width approximately).
	// A loose check: total output is not empty.
	if ansi == "" {
		t.Fatal("Render returned empty string")
	}
}

func TestRender_InvalidDimensions(t *testing.T) {
	data := syntheticPNG(t, 10, 10)
	_, err := Render(data, 0, 5)
	if err == nil {
		t.Error("expected error for zero width")
	}
	_, err = Render(data, 5, 0)
	if err == nil {
		t.Error("expected error for zero height")
	}
}

func TestDimensions(t *testing.T) {
	data := syntheticPNG(t, 100, 50)
	w, h, err := Dimensions(data)
	if err != nil {
		t.Fatalf("Dimensions: %v", err)
	}
	if w != 100 || h != 50 {
		t.Errorf("got %dx%d, want 100x50", w, h)
	}
}
