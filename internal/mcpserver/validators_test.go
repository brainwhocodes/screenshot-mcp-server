package mcpserver

import (
	"math"
	"testing"
)

func TestDefaultThreshold(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		fallback float64
		want     float64
	}{
		{"zero uses fallback", 0, 0.95, 0.95},
		{"non-zero preserved", 0.75, 0.95, 0.75},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := defaultThreshold(tt.value, tt.fallback); got != tt.want {
				t.Fatalf("defaultThreshold(%v, %v) = %v, want %v", tt.value, tt.fallback, got, tt.want)
			}
		})
	}
}

func TestValidateThreshold(t *testing.T) {
	tests := []struct {
		name    string
		value   float64
		wantErr bool
	}{
		{"low", -0.01, true},
		{"low inclusive", 0, false},
		{"high", 1.0, false},
		{"too high", 1.01, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateThreshold(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateThreshold(%v) err = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestResolveTimeoutAndPoll(t *testing.T) {
	gotTimeout, gotPoll := resolveTimeoutAndPoll(0, 0)
	if gotTimeout != defaultImageWaitTimeoutMs || gotPoll != defaultImageWaitPollMs {
		t.Fatalf("resolveTimeoutAndPoll(0, 0) = (%d, %d), want (%d, %d)", gotTimeout, gotPoll, defaultImageWaitTimeoutMs, defaultImageWaitPollMs)
	}

	gotTimeout, gotPoll = resolveTimeoutAndPoll(2000, 200)
	if gotTimeout != 2000 || gotPoll != 200 {
		t.Fatalf("resolveTimeoutAndPoll(2000, 200,...) = (%d, %d), want (2000, 200)", gotTimeout, gotPoll)
	}
}

func TestValidateWindowID(t *testing.T) {
	tests := []struct {
		name      string
		windowID  uint32
		wantError bool
	}{
		{"valid", 12345, false},
		{"zero", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWindowID(tt.windowID)
			if (err != nil) != tt.wantError {
				t.Fatalf("validateWindowID(%d) err = %v, wantErr %v", tt.windowID, err, tt.wantError)
			}
		})
	}
}

func TestValidatePositiveDimensions(t *testing.T) {
	tests := []struct {
		name      string
		width     float64
		height    float64
		wantError bool
	}{
		{"valid", 10, 20, false},
		{"zero width", 0, 20, true},
		{"zero height", 10, 0, true},
		{"negative width", -1, 20, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePositiveDimensions(tt.width, tt.height)
			if (err != nil) != tt.wantError {
				t.Fatalf("validatePositiveDimensions(%.1f, %.1f) err = %v, wantErr %v", tt.width, tt.height, err, tt.wantError)
			}
		})
	}
}

func TestNormalizeCoordSpace(t *testing.T) {
	tests := []struct {
		name      string
		coord     string
		want      string
		wantError bool
	}{
		{"default points", "", "points", false},
		{"explicit points", "points", "points", false},
		{"explicit pixels", "pixels", "pixels", false},
		{"invalid", "screen", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeCoordSpace(tt.coord)
			if (err != nil) != tt.wantError {
				t.Fatalf("normalizeCoordSpace(%q) err = %v, wantErr %v", tt.coord, err, tt.wantError)
			}
			if !tt.wantError && got != tt.want {
				t.Fatalf("normalizeCoordSpace(%q) = %q, want %q", tt.coord, got, tt.want)
			}
		})
	}
}

func TestValidateRegionInput(t *testing.T) {
	tests := []struct {
		name      string
		width     float64
		height    float64
		coord     string
		wantError bool
	}{
		{"valid", 10, 20, "points", false},
		{"default coord", 10, 20, "", false},
		{"invalid coord", 10, 20, "foo", true},
		{"bad width", 0, 20, "points", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRegionInput(tt.width, tt.height, tt.coord)
			if (err != nil) != tt.wantError {
				t.Fatalf("validateRegionInput(%.1f, %.1f, %q) err = %v, wantErr %v", tt.width, tt.height, tt.coord, err, tt.wantError)
			}
		})
	}
}

func TestValidateMaskRegions(t *testing.T) {
	tests := []struct {
		name      string
		regions   []MaskRegion
		wantError bool
	}{
		{"valid", []MaskRegion{
			{Width: 10, Height: 20},
		}, false},
		{"negative width", []MaskRegion{
			{Width: -1, Height: 20},
		}, true},
		{"negative height", []MaskRegion{
			{Width: 10, Height: -2},
		}, true},
		{"mixed", []MaskRegion{
			{Width: 10, Height: 20},
			{Width: 0, Height: 1},
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMaskRegions(tt.regions)
			if (err != nil) != tt.wantError {
				t.Fatalf("validateMaskRegions(%v) err = %v, wantErr %v", tt.regions, err, tt.wantError)
			}
		})
	}
}

func TestSafeFloatToInt(t *testing.T) {
	tests := []struct {
		name   string
		value  float64
		want   int
		wantOk bool
	}{
		{"valid", 42.75, 42, true},
		{"negative", -5, -5, true},
		{"nan", math.NaN(), 0, false},
		{"inf", math.Inf(1), 0, false},
		{"too large", float64(int(^uint(0)>>1)) + 4096, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := safeFloatToInt(tt.value)
			if got != tt.want || ok != tt.wantOk {
				t.Fatalf("safeFloatToInt(%v) = (%d, %v), want (%d, %v)", tt.value, got, ok, tt.want, tt.wantOk)
			}
		})
	}
}
