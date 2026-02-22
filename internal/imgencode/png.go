package imgencode

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
)

// EncodePNG encodes img to PNG format (lossless).
func EncodePNG(img image.Image) ([]byte, error) {
	if img == nil {
		return nil, fmt.Errorf("image is nil")
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encode png: %w", err)
	}
	return buf.Bytes(), nil
}
