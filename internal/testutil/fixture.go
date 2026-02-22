// Package testutil contains deterministic fixture helpers for integration tests.
package testutil

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"testing"
)

// WriteFixtureJPEG writes a deterministic JPEG file for integration tests.
func WriteFixtureJPEG(t *testing.T) string {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := byte(0); y < 32; y++ {
		for x := byte(0); x < 32; x++ {
			img.SetRGBA(int(x), int(y), color.RGBA{
				R: x * 8,
				G: y * 8,
				B: (x + y) * 4,
				A: 255,
			})
		}
	}

	file, err := os.CreateTemp(t.TempDir(), "fixture-*.jpg")
	if err != nil {
		t.Fatalf("create fixture file: %v", err)
	}
	defer func() {
		_ = file.Close()
	}()

	if err := jpeg.Encode(file, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("encode fixture jpeg: %v", err)
	}

	return file.Name()
}
