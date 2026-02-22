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
