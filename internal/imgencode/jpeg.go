// Package imgencode provides helpers for encoding screenshots and window images.
package imgencode

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
)

// Options controls JPEG output behavior.
type Options struct {
	Quality     int
	MaxBytes    int
	MinQuality  int
	QualityStep int
}

// DefaultOptions mirrors the Python implementation quality and size intent.
var DefaultOptions = Options{
	Quality:     60,
	MaxBytes:    1_000_000,
	MinQuality:  30,
	QualityStep: 5,
}

// EncodeJPEG encodes img to JPEG using a quality fallback loop when MaxBytes is set.
func EncodeJPEG(img image.Image, opts Options) ([]byte, error) {
	if img == nil {
		return nil, fmt.Errorf("image is nil")
	}

	opts = normalizeOptions(opts)
	quality := opts.Quality

	for {
		data, err := encodeAtQuality(img, quality)
		if err != nil {
			return nil, err
		}

		if opts.MaxBytes <= 0 || len(data) <= opts.MaxBytes || quality <= opts.MinQuality {
			return data, nil
		}

		quality -= opts.QualityStep
		if quality < opts.MinQuality {
			quality = opts.MinQuality
		}
	}
}

func normalizeOptions(opts Options) Options {
	if opts.Quality <= 0 {
		opts.Quality = DefaultOptions.Quality
	}
	if opts.MaxBytes < 0 {
		opts.MaxBytes = 0
	}
	if opts.MinQuality <= 0 {
		opts.MinQuality = DefaultOptions.MinQuality
	}
	if opts.QualityStep <= 0 {
		opts.QualityStep = DefaultOptions.QualityStep
	}
	if opts.Quality < opts.MinQuality {
		opts.Quality = opts.MinQuality
	}
	return opts
}

func encodeAtQuality(img image.Image, quality int) ([]byte, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, fmt.Errorf("encode jpeg at quality %d: %w", quality, err)
	}
	return buf.Bytes(), nil
}
