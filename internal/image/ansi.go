package image

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	_ "image/jpeg"
	"image/png"
	"time"

	"github.com/eliukblau/pixterm/pkg/ansimage"
)

// Render decodes imgBytes (JPEG or PNG) and returns an ANSI half-block string
// sized to fit within cols columns and rows terminal rows.
// Each terminal cell displays two vertical pixels (lower-half-block ▄ fg + bg color),
// so the source image is scaled to cols×(rows*2) pixels.
func Render(imgBytes []byte, cols, rows int) (string, error) {
	if cols <= 0 || rows <= 0 {
		return "", fmt.Errorf("invalid dimensions: %dx%d", cols, rows)
	}

	img, err := ansimage.NewScaledFromReader(
		bytes.NewReader(imgBytes),
		rows*2,
		cols,
		color.RGBA{0, 0, 0, 255},
		ansimage.ScaleModeFit,
		ansimage.NoDithering,
	)
	if err != nil {
		return "", fmt.Errorf("rendering ANSI image: %w", err)
	}

	return img.Render(), nil
}

// RenderGIFFrames decodes an animated GIF and returns one ANSI string per frame
// plus the per-frame display duration. Works for single-frame GIFs too.
// Delay values below 50ms are clamped to 50ms to avoid terminal thrashing.
func RenderGIFFrames(raw []byte, cols, rows int) ([]string, []time.Duration, error) {
	if cols <= 0 || rows <= 0 {
		return nil, nil, fmt.Errorf("invalid dimensions: %dx%d", cols, rows)
	}
	g, err := gif.DecodeAll(bytes.NewReader(raw))
	if err != nil {
		return nil, nil, fmt.Errorf("decoding GIF: %w", err)
	}

	canvas := image.NewRGBA(image.Rect(0, 0, g.Config.Width, g.Config.Height))
	frames := make([]string, len(g.Image))
	delays := make([]time.Duration, len(g.Image))

	for i, frame := range g.Image {
		draw.Draw(canvas, frame.Bounds(), frame, frame.Bounds().Min, draw.Over)

		var buf bytes.Buffer
		if err := png.Encode(&buf, canvas); err != nil {
			return nil, nil, fmt.Errorf("encoding GIF frame %d: %w", i, err)
		}
		ansi, err := Render(buf.Bytes(), cols, rows)
		if err != nil {
			return nil, nil, fmt.Errorf("rendering GIF frame %d: %w", i, err)
		}
		frames[i] = ansi

		d := time.Duration(g.Delay[i]) * 10 * time.Millisecond
		if d < 50*time.Millisecond {
			d = 50 * time.Millisecond
		}
		delays[i] = d
	}

	return frames, delays, nil
}

// Dimensions returns the width and height of an encoded image.
func Dimensions(imgBytes []byte) (int, int, error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(imgBytes))
	if err != nil {
		return 0, 0, err
	}
	return cfg.Width, cfg.Height, nil
}
