package mcpserver

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
)

// computeImageHash generates a hash of the image for change detection.
func computeImageHash(img image.Image, algorithm string) (string, error) {
	switch algorithm {
	case "sha256":
		return computeSHA256Hash(img)
	case "perceptual", "":
		return computePerceptualHash(img)
	default:
		return "", fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}
}

// computeSHA256Hash computes a SHA256 hash of the raw image bytes.
func computeSHA256Hash(img image.Image) (string, error) {
	bounds := img.Bounds()
	h := sha256.New()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			_, _ = h.Write([]byte{byte(r >> 8), byte(g >> 8), byte(b >> 8), byte(a >> 8)})
		}
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// computePerceptualHash computes a simple perceptual hash (average hash).
func computePerceptualHash(img image.Image) (string, error) {
	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y
	if width <= 0 || height <= 0 {
		return "", fmt.Errorf("invalid image dimensions")
	}

	sampleSize := 8
	var sum uint64
	var count uint64
	pixels := make([]uint8, sampleSize*sampleSize)

	for y := 0; y < sampleSize; y++ {
		for x := 0; x < sampleSize; x++ {
			srcX := bounds.Min.X + (x * width / sampleSize)
			srcY := bounds.Min.Y + (y * height / sampleSize)

			r, g, b, _ := img.At(srcX, srcY).RGBA()
			gray := uint8((0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 256)
			pixels[y*sampleSize+x] = gray
			sum += uint64(gray)
			count++
		}
	}

	if count == 0 {
		return "", fmt.Errorf("invalid template image for perceptual hash")
	}
	avgRaw := sum / count
	if avgRaw > uint64(^uint8(0)) {
		return "", fmt.Errorf("perceptual hash average overflow: %d", avgRaw)
	}
	avg := uint8(avgRaw)

	var hash uint64
	for i, pixel := range pixels {
		if i >= 64 {
			break
		}
		hash <<= 1
		if pixel > avg {
			hash |= 1
		}
	}

	return fmt.Sprintf("%016x", hash), nil
}

// Point represents a coordinate.
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// ImageMatch represents a found template match.
type ImageMatch struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Score  float64 `json:"score"`
}

// ImageComparisonResult represents the result of comparing two images.
type ImageComparisonResult struct {
	Similarity  float64 `json:"similarity"`
	Match       bool    `json:"match"`
	DiffPixels  int     `json:"diff_pixels"`
	TotalPixels int     `json:"total_pixels"`
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func downscaleColorChannel(value uint32) uint8 {
	scaled := (value >> 8) & 0xff
	if scaled > 0xff {
		return 0xff
	}
	return uint8(scaled)
}

func toRGBA8(c color.Color) color.RGBA {
	r, g, b, a := c.RGBA()
	return color.RGBA{
		R: downscaleColorChannel(r),
		G: downscaleColorChannel(g),
		B: downscaleColorChannel(b),
		A: downscaleColorChannel(a),
	}
}
