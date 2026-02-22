package window

import "testing"

func TestWindow_IsTiny(t *testing.T) {
	tests := []struct {
		name   string
		width  float64
		height float64
		want   bool
	}{
		{"normal window", 800, 600, false},
		{"exactly 50x50", 50, 50, false},
		{"tiny width", 49, 600, true},
		{"tiny height", 800, 49, true},
		{"both tiny", 40, 40, true},
		{"large window", 1920, 1080, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := Window{Bounds: Bounds{Width: tt.width, Height: tt.height}}
			if got := w.IsTiny(); got != tt.want {
				t.Errorf("IsTiny() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWindow_IsSystemWindow(t *testing.T) {
	tests := []struct {
		name      string
		ownerName string
		want      bool
	}{
		{"Window Server", "Window Server", true},
		{"SystemUIServer", "SystemUIServer", true},
		{"Dock", "Dock", true},
		{"loginwindow", "loginwindow", true},
		{"Finder", "Finder", false},
		{"Safari", "Safari", false},
		{"empty owner", "", true},
		{"random app", "MyApp", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := Window{OwnerName: tt.ownerName}
			if got := w.IsSystemWindow(); got != tt.want {
				t.Errorf("IsSystemWindow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBounds_ContainsPoint(t *testing.T) {
	b := Bounds{X: 100, Y: 100, Width: 800, Height: 600}

	tests := []struct {
		name string
		x    float64
		y    float64
		want bool
	}{
		{"inside top-left", 100, 100, true},
		{"inside center", 500, 400, true},
		{"inside bottom-right", 899, 699, true},
		{"outside left", 99, 400, false},
		{"outside right", 900, 400, false},
		{"outside top", 500, 99, false},
		{"outside bottom", 500, 700, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contains := tt.x >= b.X && tt.x < b.X+b.Width &&
				tt.y >= b.Y && tt.y < b.Y+b.Height
			if contains != tt.want {
				t.Errorf("point (%.0f, %.0f) contains = %v, want %v", tt.x, tt.y, contains, tt.want)
			}
		})
	}
}

func TestCoordinateMapping_PixelToPoint(t *testing.T) {
	bounds := Bounds{X: 100, Y: 50, Width: 400, Height: 300}
	scale := 2.0

	tests := []struct {
		name    string
		pxX     float64
		pxY     float64
		wantPtX float64
		wantPtY float64
	}{
		{"top-left pixel", 0, 0, 100, 50},
		{"center pixel", 400, 300, 300, 200},
		{"bottom-right pixel", 798, 598, 499, 349},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptX := bounds.X + (tt.pxX / scale)
			ptY := bounds.Y + (tt.pxY / scale)

			if ptX != tt.wantPtX {
				t.Errorf("ptX = %.0f, want %.0f", ptX, tt.wantPtX)
			}
			if ptY != tt.wantPtY {
				t.Errorf("ptY = %.0f, want %.0f", ptY, tt.wantPtY)
			}
		})
	}
}

func TestCoordinateClamp(t *testing.T) {
	imgWidth := 800
	imgHeight := 600

	tests := []struct {
		name  string
		x     float64
		y     float64
		wantX float64
		wantY float64
	}{
		{"normal", 400, 300, 400, 300},
		{"negative x", -10, 300, 0, 300},
		{"negative y", 400, -10, 400, 0},
		{"over max x", 900, 300, 799, 300},
		{"over max y", 400, 700, 400, 599},
		{"both over", 900, 700, 799, 599},
		{"both negative", -10, -10, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y := tt.x, tt.y

			if x < 0 {
				x = 0
			}
			if y < 0 {
				y = 0
			}
			if x >= float64(imgWidth) {
				x = float64(imgWidth) - 1
			}
			if y >= float64(imgHeight) {
				y = float64(imgHeight) - 1
			}

			if x != tt.wantX {
				t.Errorf("x = %.0f, want %.0f", x, tt.wantX)
			}
			if y != tt.wantY {
				t.Errorf("y = %.0f, want %.0f", y, tt.wantY)
			}
		})
	}
}
