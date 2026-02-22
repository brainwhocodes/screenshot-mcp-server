// Package testutil contains deterministic fixture helpers for integration tests.
package testutil

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

// fixtureSize defines the deterministic fixture image dimensions.
const fixtureSize = 32

var fixtureSeq uint64

// WriteFixtureJPEG writes a deterministic JPEG fixture for integration tests.
func WriteFixtureJPEG(t *testing.T) string {
	t.Helper()
	return writeFixtureImage(t, "fixture", ".jpg", encodeFixtureJPEG)
}

// WriteFixturePNG writes a deterministic PNG fixture for integration tests.
func WriteFixturePNG(t *testing.T) string {
	t.Helper()
	return writeFixtureImage(t, "fixture", ".png", encodeFixturePNG)
}

func writeFixtureImage(t *testing.T, prefix, ext string, encode func(*testing.T, string, image.Image) string) string {
	t.Helper()
	seq := atomic.AddUint64(&fixtureSeq, 1)

	// Keep fixture generation deterministic for hash/comparison tests.
	img := image.NewRGBA(image.Rect(0, 0, fixtureSize, fixtureSize))
	for y := byte(0); y < fixtureSize; y++ {
		for x := byte(0); x < fixtureSize; x++ {
			img.SetRGBA(int(x), int(y), color.RGBA{
				R: x * 8,
				G: y * 8,
				B: (x + y) * 4,
				A: 255,
			})
		}
	}

	path := filepath.Join(t.TempDir(), fmt.Sprintf("%s-%06d%s", prefix, seq, ext))
	return encode(t, path, img)
}

func encodeFixtureJPEG(t *testing.T, path string, img image.Image) string {
	t.Helper()
	// Accepted G304 suppression: fixture path is deterministic and test-controlled.
	// #nosec G304
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create fixture file: %v", err)
	}
	defer func() {
		_ = file.Close()
	}()

	if err := jpeg.Encode(file, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("encode fixture jpeg: %v", err)
	}
	return path
}

func encodeFixturePNG(t *testing.T, path string, img image.Image) string {
	t.Helper()
	// Accepted G304 suppression: fixture path is deterministic and test-controlled.
	// #nosec G304
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create fixture file: %v", err)
	}
	defer func() {
		_ = file.Close()
	}()

	if err := png.Encode(file, img); err != nil {
		t.Fatalf("encode fixture png: %v", err)
	}
	return path
}
