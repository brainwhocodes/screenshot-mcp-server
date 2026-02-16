package imgencode

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"math/rand"
	"testing"
)

func TestEncodeJPEG_ValidJPEG(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.SetRGBA(x, y, color.RGBA{
				R: uint8(x * 4),
				G: uint8(y * 4),
				B: uint8((x + y) * 2),
				A: 255,
			})
		}
	}

	data, err := EncodeJPEG(img, Options{Quality: 60})
	if err != nil {
		t.Fatalf("EncodeJPEG failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty JPEG data")
	}

	if _, err := jpeg.Decode(bytes.NewReader(data)); err != nil {
		t.Fatalf("decode jpeg: %v", err)
	}
}

func TestEncodeJPEG_SizeReductionWithMaxBytes(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 512, 512))
	rng := rand.New(rand.NewSource(42))
	for y := 0; y < 512; y++ {
		for x := 0; x < 512; x++ {
			img.SetRGBA(x, y, color.RGBA{
				R: uint8(rng.Intn(256)),
				G: uint8(rng.Intn(256)),
				B: uint8(rng.Intn(256)),
				A: 255,
			})
		}
	}

	baseline, err := EncodeJPEG(img, Options{Quality: 95})
	if err != nil {
		t.Fatalf("baseline encode failed: %v", err)
	}

	limited, err := EncodeJPEG(img, Options{
		Quality:     95,
		MaxBytes:    len(baseline) / 3,
		MinQuality:  20,
		QualityStep: 5,
	})
	if err != nil {
		t.Fatalf("limited encode failed: %v", err)
	}
	if len(limited) > len(baseline) {
		t.Fatalf("expected limited jpeg to be <= baseline size: limited=%d baseline=%d", len(limited), len(baseline))
	}

	if _, err := jpeg.Decode(bytes.NewReader(limited)); err != nil {
		t.Fatalf("decode limited jpeg: %v", err)
	}
}

func TestEncodeJPEG_NilImage(t *testing.T) {
	if _, err := EncodeJPEG(nil, DefaultOptions); err == nil {
		t.Fatal("expected error for nil image")
	}
}
