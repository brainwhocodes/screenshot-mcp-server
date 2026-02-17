package window

/*
#cgo LDFLAGS: -framework CoreGraphics -framework CoreFoundation -framework ApplicationServices
#include <CoreGraphics/CoreGraphics.h>
#include <CoreFoundation/CoreFoundation.h>
#include <ApplicationServices/ApplicationServices.h>
#include <stdlib.h>
#include <string.h>

// Helper to check if CFDictionaryRef is NULL
int is_dict_null(CFDictionaryRef dict) {
    return dict == NULL ? 1 : 0;
}

// Helper to check if CFArrayRef is NULL
int is_array_null(CFArrayRef arr) {
    return arr == NULL ? 1 : 0;
}

// Helper to check if CFStringRef is NULL
int is_string_null(CFStringRef str) {
    return str == NULL ? 1 : 0;
}

// Helper to check if CFNumberRef is NULL
int is_number_null(CFNumberRef num) {
    return num == NULL ? 1 : 0;
}

// Helper to get CFDictionary values
CFStringRef cf_dict_get_string(CFDictionaryRef dict, const char *key) {
    CFStringRef cfKey = CFStringCreateWithCString(kCFAllocatorDefault, key, kCFStringEncodingUTF8);
    if (cfKey == NULL) return NULL;
    CFStringRef value = (CFStringRef)CFDictionaryGetValue(dict, cfKey);
    CFRelease(cfKey);
    return value;
}

CFNumberRef cf_dict_get_number(CFDictionaryRef dict, const char *key) {
    CFStringRef cfKey = CFStringCreateWithCString(kCFAllocatorDefault, key, kCFStringEncodingUTF8);
    if (cfKey == NULL) return NULL;
    CFNumberRef value = (CFNumberRef)CFDictionaryGetValue(dict, cfKey);
    CFRelease(cfKey);
    return value;
}

CFDictionaryRef cf_dict_get_dict(CFDictionaryRef dict, const char *key) {
    CFStringRef cfKey = CFStringCreateWithCString(kCFAllocatorDefault, key, kCFStringEncodingUTF8);
    if (cfKey == NULL) return NULL;
    CFDictionaryRef value = (CFDictionaryRef)CFDictionaryGetValue(dict, cfKey);
    CFRelease(cfKey);
    return value;
}

// Convert CFString to C string - caller must free
char* cfstring_to_cstring(CFStringRef str) {
    if (str == NULL) return NULL;
    CFIndex len = CFStringGetLength(str);
    CFIndex maxLen = CFStringGetMaximumSizeForEncoding(len, kCFStringEncodingUTF8) + 1;
    char *buffer = (char*)malloc(maxLen);
    if (buffer == NULL) return NULL;
    if (!CFStringGetCString(str, buffer, maxLen, kCFStringEncodingUTF8)) {
        free(buffer);
        return NULL;
    }
    return buffer;
}

// Get int32 from CFNumber
int32_t cfnumber_to_int32(CFNumberRef num) {
    int32_t value = 0;
    if (num != NULL) {
        CFNumberGetValue(num, kCFNumberIntType, &value);
    }
    return value;
}

// Get double from CFNumber
double cfnumber_to_double(CFNumberRef num) {
    double value = 0;
    if (num != NULL) {
        CFNumberGetValue(num, kCFNumberDoubleType, &value);
    }
    return value;
}

// Post mouse click event
void post_mouse_click(double x, double y, int button, int clicks) {
    CGEventType downType, upType;
    CGMouseButton mouseButton;

    if (button == 0) { // left
        downType = kCGEventLeftMouseDown;
        upType = kCGEventLeftMouseUp;
        mouseButton = kCGMouseButtonLeft;
    } else if (button == 1) { // right
        downType = kCGEventRightMouseDown;
        upType = kCGEventRightMouseUp;
        mouseButton = kCGMouseButtonRight;
    } else {
        downType = kCGEventLeftMouseDown;
        upType = kCGEventLeftMouseUp;
        mouseButton = kCGMouseButtonLeft;
    }

    CGPoint point = CGPointMake(x, y);

    for (int i = 0; i < clicks; i++) {
        CGEventRef downEvent = CGEventCreateMouseEvent(NULL, downType, point, mouseButton);
        if (downEvent) {
            CGEventPost(kCGHIDEventTap, downEvent);
            CFRelease(downEvent);
        }

        CGEventRef upEvent = CGEventCreateMouseEvent(NULL, upType, point, mouseButton);
        if (upEvent) {
            CGEventPost(kCGHIDEventTap, upEvent);
            CFRelease(upEvent);
        }
    }
}

// Check screen capture permission
int has_screen_capture_access() {
    return CGPreflightScreenCaptureAccess() ? 1 : 0;
}

// Check accessibility permission
int has_accessibility_access() {
    return AXIsProcessTrusted() ? 1 : 0;
}
*/
import "C"
import (
	"context"
	"fmt"
	"image"
	"os/exec"
	"unsafe"

	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/imgencode"
	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/screenshot"
)

// Window represents a macOS window
type Window struct {
	WindowID   uint32 `json:"window_id"`
	OwnerName  string `json:"owner_name"`
	PID        int32  `json:"pid"`
	Title      string `json:"title"`
	Bounds     Bounds `json:"bounds"`
	IsOnScreen bool   `json:"is_on_screen"`
}

// Bounds represents window bounds in screen coordinates (points)
type Bounds struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// ScreenshotMetadata contains metadata about a window screenshot
type ScreenshotMetadata struct {
	WindowID    uint32  `json:"window_id"`
	Bounds      Bounds  `json:"bounds"`
	ImageWidth  int     `json:"image_width"`
	ImageHeight int     `json:"image_height"`
	Scale       float64 `json:"scale"`
}

// IsTiny returns true if window is smaller than 50x50
func (w *Window) IsTiny() bool {
	return w.Bounds.Width < 50 || w.Bounds.Height < 50
}

// IsSystemWindow returns true if this looks like a system overlay window
func (w *Window) IsSystemWindow() bool {
	// Filter out known system/window server windows
	systemOwners := map[string]bool{
		"Window Server":  true,
		"SystemUIServer": true,
		"Dock":           true,
		"loginwindow":    true,
		"coreauthd":      true,
		"AppleSpell":     true,
		"Finder":         false, // Keep Finder
		"":               true,
	}

	if isSystem, exists := systemOwners[w.OwnerName]; exists {
		return isSystem
	}

	return false
}

// ListWindows returns all visible windows
func ListWindows(ctx context.Context) ([]Window, error) {
	// Get window list with bounds
	windowList := C.CGWindowListCopyWindowInfo(
		C.kCGWindowListOptionOnScreenOnly|C.kCGWindowListExcludeDesktopElements,
		C.kCGNullWindowID,
	)
	if C.is_array_null(windowList) == 1 {
		return nil, fmt.Errorf("failed to get window list")
	}
	defer C.CFRelease(C.CFTypeRef(windowList))

	count := C.CFArrayGetCount(windowList)
	var windows []Window

	for i := C.CFIndex(0); i < count; i++ {
		windowRef := C.CFArrayGetValueAtIndex(windowList, i)
		windowDict := C.CFDictionaryRef(windowRef)

		win := parseWindowInfo(windowDict)
		if win != nil && !win.IsTiny() && !win.IsSystemWindow() && win.IsOnScreen {
			windows = append(windows, *win)
		}
	}

	return windows, nil
}

func parseWindowInfo(dict C.CFDictionaryRef) *Window {
	if C.is_dict_null(dict) == 1 {
		return nil
	}

	win := &Window{}

	// Get window ID (kCGWindowNumber)
	cKeyWindowNumber := C.CString("kCGWindowNumber")
	if num := C.cf_dict_get_number(dict, cKeyWindowNumber); C.is_number_null(num) == 0 {
		win.WindowID = uint32(int32(C.cfnumber_to_int32(num)))
	}
	C.free(unsafe.Pointer(cKeyWindowNumber))

	// Get owner name (kCGWindowOwnerName)
	cKeyOwnerName := C.CString("kCGWindowOwnerName")
	if str := C.cf_dict_get_string(dict, cKeyOwnerName); C.is_string_null(str) == 0 {
		cstr := C.cfstring_to_cstring(str)
		if cstr != nil {
			win.OwnerName = C.GoString(cstr)
			C.free(unsafe.Pointer(cstr))
		}
	}
	C.free(unsafe.Pointer(cKeyOwnerName))

	// Get PID (kCGWindowOwnerPID)
	cKeyOwnerPID := C.CString("kCGWindowOwnerPID")
	if num := C.cf_dict_get_number(dict, cKeyOwnerPID); C.is_number_null(num) == 0 {
		win.PID = int32(C.cfnumber_to_int32(num))
	}
	C.free(unsafe.Pointer(cKeyOwnerPID))

	// Get title (kCGWindowName)
	cKeyWindowName := C.CString("kCGWindowName")
	if str := C.cf_dict_get_string(dict, cKeyWindowName); C.is_string_null(str) == 0 {
		cstr := C.cfstring_to_cstring(str)
		if cstr != nil {
			win.Title = C.GoString(cstr)
			C.free(unsafe.Pointer(cstr))
		}
	}
	C.free(unsafe.Pointer(cKeyWindowName))

	// Get bounds (kCGWindowBounds)
	cKeyBounds := C.CString("kCGWindowBounds")
	if boundsDict := C.cf_dict_get_dict(dict, cKeyBounds); C.is_dict_null(boundsDict) == 0 {
		win.Bounds = parseBounds(boundsDict)
	}
	C.free(unsafe.Pointer(cKeyBounds))

	// Check if on screen (kCGWindowIsOnscreen)
	cKeyOnscreen := C.CString("kCGWindowIsOnscreen")
	if num := C.cf_dict_get_number(dict, cKeyOnscreen); C.is_number_null(num) == 0 {
		win.IsOnScreen = int32(C.cfnumber_to_int32(num)) != 0
	} else {
		win.IsOnScreen = true // Default to true if not present
	}
	C.free(unsafe.Pointer(cKeyOnscreen))

	return win
}

func parseBounds(dict C.CFDictionaryRef) Bounds {
	b := Bounds{}

	cKeyX := C.CString("X")
	if x := C.cf_dict_get_number(dict, cKeyX); C.is_number_null(x) == 0 {
		b.X = float64(float64(C.cfnumber_to_double(x)))
	}
	C.free(unsafe.Pointer(cKeyX))

	cKeyY := C.CString("Y")
	if y := C.cf_dict_get_number(dict, cKeyY); C.is_number_null(y) == 0 {
		b.Y = float64(float64(C.cfnumber_to_double(y)))
	}
	C.free(unsafe.Pointer(cKeyY))

	cKeyWidth := C.CString("Width")
	if w := C.cf_dict_get_number(dict, cKeyWidth); C.is_number_null(w) == 0 {
		b.Width = float64(float64(C.cfnumber_to_double(w)))
	}
	C.free(unsafe.Pointer(cKeyWidth))

	cKeyHeight := C.CString("Height")
	if h := C.cf_dict_get_number(dict, cKeyHeight); C.is_number_null(h) == 0 {
		b.Height = float64(float64(C.cfnumber_to_double(h)))
	}
	C.free(unsafe.Pointer(cKeyHeight))

	return b
}

// FocusWindow brings a window to the foreground
func FocusWindow(ctx context.Context, windowID uint32) error {
	// Get the window info to find its PID
	windows, err := ListWindows(ctx)
	if err != nil {
		return fmt.Errorf("list windows: %w", err)
	}

	var targetWindow *Window
	for i := range windows {
		if windows[i].WindowID == windowID {
			targetWindow = &windows[i]
			break
		}
	}

	if targetWindow == nil {
		return fmt.Errorf("window %d not found", windowID)
	}

	// Use AppleScript to activate the application
	script := fmt.Sprintf(`tell application "System Events" to tell application process "%s" to set frontmost to true`, targetWindow.OwnerName)
	return runAppleScript(script)
}

func runAppleScript(script string) error {
	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("apple script failed: %w, output: %s", err, string(output))
	}
	return nil
}

// TakeWindowScreenshot captures a window and returns JPEG bytes with metadata
// Uses fallback: take full screen screenshot and crop to window bounds
func TakeWindowScreenshot(ctx context.Context, windowID uint32, opts imgencode.Options) ([]byte, *ScreenshotMetadata, error) {
	// Get current window info
	windows, err := ListWindows(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("list windows: %w", err)
	}

	var targetWindow *Window
	for i := range windows {
		if windows[i].WindowID == windowID {
			targetWindow = &windows[i]
			break
		}
	}

	if targetWindow == nil {
		return nil, nil, fmt.Errorf("window %d not found", windowID)
	}

	// Capture full screen using existing capturer
	capturer := screenshot.NewCapturer()
	fullImg, err := capturer.Capture(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("capture screen: %w", err)
	}

	// Calculate pixel bounds from points
	// For Retina displays, we need to account for scale factor
	bounds := targetWindow.Bounds

	// Get the scale factor by comparing image bounds to screen bounds
	// This is a rough approximation - for accurate results we'd need to query the display
	// But for now we'll assume uniform scaling
	imgBounds := fullImg.Bounds()
	scaleX := float64(imgBounds.Dx()) / 2560.0 // Assuming standard screen width in points
	scaleY := float64(imgBounds.Dy()) / 1440.0 // Assuming standard screen height in points

	// Use the average scale
	scale := (scaleX + scaleY) / 2.0
	if scale < 1.0 {
		scale = 1.0
	}

	// Convert points to pixels
	x1 := int(bounds.X * scale)
	y1 := int(bounds.Y * scale)
	x2 := int((bounds.X + bounds.Width) * scale)
	y2 := int((bounds.Y + bounds.Height) * scale)

	// Ensure bounds are within image
	if x1 < 0 {
		x1 = 0
	}
	if y1 < 0 {
		y1 = 0
	}
	if x2 > imgBounds.Max.X {
		x2 = imgBounds.Max.X
	}
	if y2 > imgBounds.Max.Y {
		y2 = imgBounds.Max.Y
	}

	// Crop the image
	cropRect := image.Rect(x1, y1, x2, y2)
	croppedImg := fullImg.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(cropRect)

	// Encode to JPEG
	data, err := imgencode.EncodeJPEG(croppedImg, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("encode screenshot: %w", err)
	}

	// Calculate actual dimensions
	imgWidth := x2 - x1
	imgHeight := y2 - y1

	metadata := &ScreenshotMetadata{
		WindowID:    windowID,
		Bounds:      bounds,
		ImageWidth:  imgWidth,
		ImageHeight: imgHeight,
		Scale:       scale,
	}

	return data, metadata, nil
}

// Click performs a mouse click at the specified coordinates
// x, y are pixel coordinates in the screenshot image
func Click(ctx context.Context, windowID uint32, x, y float64, button string, clicks int) error {
	// Get current window info
	windows, err := ListWindows(ctx)
	if err != nil {
		return fmt.Errorf("list windows: %w", err)
	}

	var targetWindow *Window
	for i := range windows {
		if windows[i].WindowID == windowID {
			targetWindow = &windows[i]
			break
		}
	}

	if targetWindow == nil {
		return fmt.Errorf("window %d not found", windowID)
	}

	// Take a fresh screenshot to get current scale
	_, metadata, err := TakeWindowScreenshot(ctx, windowID, imgencode.Options{Quality: 1})
	if err != nil {
		return fmt.Errorf("take screenshot for coordinate mapping: %w", err)
	}

	// Clamp coordinates to image bounds
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x >= float64(metadata.ImageWidth) {
		x = float64(metadata.ImageWidth) - 1
	}
	if y >= float64(metadata.ImageHeight) {
		y = float64(metadata.ImageHeight) - 1
	}

	// Convert pixel coordinates to screen points
	xPt := targetWindow.Bounds.X + (x / metadata.Scale)
	yPt := targetWindow.Bounds.Y + (y / metadata.Scale)

	// Verify point is within window bounds
	if xPt < targetWindow.Bounds.X || xPt >= targetWindow.Bounds.X+targetWindow.Bounds.Width ||
		yPt < targetWindow.Bounds.Y || yPt >= targetWindow.Bounds.Y+targetWindow.Bounds.Height {
		return fmt.Errorf("click coordinates outside window bounds")
	}

	// Determine button
	btn := 0 // left
	if button == "right" {
		btn = 1
	}

	// Post mouse event
	C.post_mouse_click(C.double(xPt), C.double(yPt), C.int(btn), C.int(clicks))

	return nil
}

// CheckPermissions checks if required permissions are granted
func CheckPermissions() (screenRecording bool, accessibility bool) {
	screenRecording = C.has_screen_capture_access() != 0
	accessibility = C.has_accessibility_access() != 0
	return
}
