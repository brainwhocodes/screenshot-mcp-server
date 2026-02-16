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
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			img.SetRGBA(x, y, color.RGBA{
				R: uint8(x * 8),
				G: uint8(y * 8),
				B: uint8((x + y) * 4),
				A: 255,
			})
		}
	}

	file, err := os.CreateTemp(t.TempDir(), "fixture-*.jpg")
	if err != nil {
		t.Fatalf("create fixture file: %v", err)
	}
	defer file.Close()

	if err := jpeg.Encode(file, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("encode fixture jpeg: %v", err)
	}

	return file.Name()
}
