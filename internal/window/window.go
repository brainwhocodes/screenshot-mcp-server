//go:build darwin
// +build darwin

// Package window provides macOS window discovery and input helpers used by MCP tools.
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

// Get display scale factor for the main display
double get_display_scale() {
    CGDirectDisplayID mainDisplay = CGMainDisplayID();
    CGDisplayModeRef mode = CGDisplayCopyDisplayMode(mainDisplay);
    if (mode == NULL) {
        return 1.0;
    }

    size_t widthPixels = CGDisplayModeGetPixelWidth(mode);
    size_t widthPoints = CGDisplayPixelsWide(mainDisplay);
    if (widthPoints == 0) {
        CGDisplayModeRelease(mode);
        return 1.0;
    }

    // Calculate scale: on Retina, pixels > points
    double scale = (double)widthPixels / (double)widthPoints;
    CGDisplayModeRelease(mode);

    if (scale < 1.0) {
        scale = 1.0;
    }
    return scale;
}

// Get scale factor for a point on screen (handles multi-monitor)
double get_scale_at_point(double x, double y) {
    CGPoint point = CGPointMake(x, y);
    CGDirectDisplayID displayID;
    uint32_t count;

    if (CGGetDisplaysWithPoint(point, 1, &displayID, &count) != kCGErrorSuccess || count == 0) {
        return get_display_scale();
    }

    CGDisplayModeRef mode = CGDisplayCopyDisplayMode(displayID);
    if (mode == NULL) {
        return 1.0;
    }

    size_t widthPixels = CGDisplayModeGetPixelWidth(mode);
    size_t widthPoints = CGDisplayPixelsWide(displayID);
    CGDisplayModeRelease(mode);

    if (widthPoints == 0) {
        return 1.0;
    }

    double scale = (double)widthPixels / (double)widthPoints;
    if (scale < 1.0) {
        scale = 1.0;
    }
    return scale;
}

// Post mouse move event
void post_mouse_move(double x, double y) {
    CGPoint point = CGPointMake(x, y);
    CGEventRef event = CGEventCreateMouseEvent(NULL, kCGEventMouseMoved, point, kCGMouseButtonLeft);
    if (event) {
        CGEventPost(kCGHIDEventTap, event);
        CFRelease(event);
    }
}

// Post mouse down event
void post_mouse_down(double x, double y, int button) {
    CGEventType downType;
    CGMouseButton mouseButton;

    if (button == 0) { // left
        downType = kCGEventLeftMouseDown;
        mouseButton = kCGMouseButtonLeft;
    } else if (button == 1) { // right
        downType = kCGEventRightMouseDown;
        mouseButton = kCGMouseButtonRight;
    } else { // middle
        downType = kCGEventOtherMouseDown;
        mouseButton = kCGMouseButtonCenter;
    }

    CGPoint point = CGPointMake(x, y);
    CGEventRef event = CGEventCreateMouseEvent(NULL, downType, point, mouseButton);
    if (event) {
        CGEventPost(kCGHIDEventTap, event);
        CFRelease(event);
    }
}

// Post mouse up event
void post_mouse_up(double x, double y, int button) {
    CGEventType upType;
    CGMouseButton mouseButton;

    if (button == 0) { // left
        upType = kCGEventLeftMouseUp;
        mouseButton = kCGMouseButtonLeft;
    } else if (button == 1) { // right
        upType = kCGEventRightMouseUp;
        mouseButton = kCGMouseButtonRight;
    } else { // middle
        upType = kCGEventOtherMouseUp;
        mouseButton = kCGMouseButtonCenter;
    }

    CGPoint point = CGPointMake(x, y);
    CGEventRef event = CGEventCreateMouseEvent(NULL, upType, point, mouseButton);
    if (event) {
        CGEventPost(kCGHIDEventTap, event);
        CFRelease(event);
    }
}

// Post scroll event (deltaX and deltaY in pixels, positive = right/down)
void post_scroll(double x, double y, double deltaX, double deltaY) {
    CGPoint point = CGPointMake(x, y);
    CGEventRef event = CGEventCreateScrollWheelEvent(NULL, kCGScrollEventUnitPixel, 2, (int32_t)deltaY, (int32_t)deltaX);
    if (event) {
        CGEventPost(kCGHIDEventTap, event);
        CFRelease(event);
    }
}
*/
import "C"
import (
	"context"
	"fmt"
	"image"
	"unsafe"

	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/imgencode"
	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/safeexec"
	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/screenshot"
)

// SupportsWindowTools reports whether this platform supports window-level automation features.
func SupportsWindowTools() bool {
	return true
}

// UnsupportedWindowToolsReason returns the human-readable reason automation features are unavailable.
func UnsupportedWindowToolsReason() string {
	return ""
}

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
func ListWindows(_ context.Context) ([]Window, error) {
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

	quotedOwnerName := safeexec.QuoteAppleScriptString(targetWindow.OwnerName)
	script := fmt.Sprintf(`tell application "System Events" to tell application process "%s" to set frontmost to true`, quotedOwnerName)
	if err := safeexec.RunAppleScript(ctx, script); err != nil {
		centerX := targetWindow.Bounds.X + targetWindow.Bounds.Width/2
		centerY := targetWindow.Bounds.Y + targetWindow.Bounds.Height/2
		C.post_mouse_click(C.double(centerX), C.double(centerY), C.int(0), C.int(1))
	}

	return nil
}

// TakeWindowScreenshot captures a window and returns JPEG bytes with metadata
// Uses full screen capture and crop to window bounds
func TakeWindowScreenshot(ctx context.Context, windowID uint32, opts imgencode.Options) ([]byte, *ScreenshotMetadata, error) {
	encode := func(img image.Image) ([]byte, error) {
		return imgencode.EncodeJPEG(img, opts)
	}
	return captureWindowImage(ctx, windowID, encode)
}

func captureWindowImage(ctx context.Context, windowID uint32, encode func(image.Image) ([]byte, error)) ([]byte, *ScreenshotMetadata, error) {
	targetWindow, err := findWindowByID(ctx, windowID)
	if err != nil {
		return nil, nil, err
	}

	capturer := screenshot.NewCapturer()
	fullImg, err := capturer.Capture(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("capture screen: %w", err)
	}

	scale := getScaleForWindow(targetWindow.Bounds)
	cropRect := cropRectForWindow(targetWindow.Bounds, fullImg.Bounds(), scale)
	croppedImg := cropImage(fullImg, cropRect)

	data, err := encode(croppedImg)
	if err != nil {
		return nil, nil, fmt.Errorf("encode screenshot: %w", err)
	}

	return data, &ScreenshotMetadata{
		WindowID:    windowID,
		Bounds:      targetWindow.Bounds,
		ImageWidth:  cropRect.Dx(),
		ImageHeight: cropRect.Dy(),
		Scale:       scale,
	}, nil
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

// RegionMetadata contains metadata about a region screenshot
type RegionMetadata struct {
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Width       float64 `json:"width"`
	Height      float64 `json:"height"`
	ImageWidth  int     `json:"image_width"`
	ImageHeight int     `json:"image_height"`
	Scale       float64 `json:"scale"`
	CoordSpace  string  `json:"coord_space"`
}

// TakeRegionScreenshot captures a region of the screen and returns JPEG bytes with metadata.
// coordSpace can be "points" (screen coordinates) or "pixels" (image coordinates).
// When coordSpace is "points", the region is specified in screen points (Quartz coordinates).
// When coordSpace is "pixels", the region is specified in pixel coordinates.
func TakeRegionScreenshot(ctx context.Context, x, y, width, height float64, coordSpace string, opts imgencode.Options) ([]byte, *RegionMetadata, error) {
	encode := func(img image.Image) ([]byte, error) {
		return imgencode.EncodeJPEG(img, opts)
	}
	return captureRegionScreenshot(ctx, x, y, width, height, coordSpace, encode)
}

// TakeWindowScreenshotPNG captures a window and returns PNG bytes (lossless).
func TakeWindowScreenshotPNG(ctx context.Context, windowID uint32) ([]byte, *ScreenshotMetadata, error) {
	encode := func(img image.Image) ([]byte, error) {
		return imgencode.EncodePNG(img)
	}
	return captureWindowImage(ctx, windowID, encode)
}

// TakeRegionScreenshotPNG captures a region and returns PNG bytes (lossless).
func TakeRegionScreenshotPNG(ctx context.Context, x, y, width, height float64, coordSpace string) ([]byte, *RegionMetadata, error) {
	encode := func(img image.Image) ([]byte, error) {
		return imgencode.EncodePNG(img)
	}
	return captureRegionScreenshot(ctx, x, y, width, height, coordSpace, encode)
}

func captureRegionScreenshot(ctx context.Context, x, y, width, height float64, coordSpace string, encode func(image.Image) ([]byte, error)) ([]byte, *RegionMetadata, error) {
	fullImg, err := screenshot.NewCapturer().Capture(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("capture screen: %w", err)
	}

	centerX := x + width/2
	centerY := y + height/2
	scale := float64(C.get_scale_at_point(C.double(centerX), C.double(centerY)))
	cropRect := cropRectForRegion(fullImg.Bounds(), x, y, width, height, scale, coordSpace)
	croppedImg := cropImage(fullImg, cropRect)

	data, err := encode(croppedImg)
	if err != nil {
		return nil, nil, fmt.Errorf("encode screenshot: %w", err)
	}

	metadata := &RegionMetadata{
		X:           float64(cropRect.Min.X) / scale,
		Y:           float64(cropRect.Min.Y) / scale,
		Width:       float64(cropRect.Dx()) / scale,
		Height:      float64(cropRect.Dy()) / scale,
		ImageWidth:  cropRect.Dx(),
		ImageHeight: cropRect.Dy(),
		Scale:       scale,
		CoordSpace:  coordSpace,
	}
	return data, metadata, nil
}

func cropRectForRegion(imgBounds image.Rectangle, x, y, width, height, scale float64, coordSpace string) image.Rectangle {
	var x1, y1, x2, y2 int
	if coordSpace == "pixels" {
		x1, y1, x2, y2 = int(x), int(y), int(x+width), int(y+height)
	} else {
		x1, y1, x2, y2 = int(x*scale), int(y*scale), int((x+width)*scale), int((y+height)*scale)
	}

	return clampRect(image.Rect(x1, y1, x2, y2), imgBounds)
}

func getScaleForWindow(bounds Bounds) float64 {
	centerX := bounds.X + bounds.Width/2
	centerY := bounds.Y + bounds.Height/2
	return float64(C.get_scale_at_point(C.double(centerX), C.double(centerY)))
}

func cropRectForWindow(bounds Bounds, imgBounds image.Rectangle, scale float64) image.Rectangle {
	x1 := int(bounds.X * scale)
	y1 := int(bounds.Y * scale)
	x2 := int((bounds.X + bounds.Width) * scale)
	y2 := int((bounds.Y + bounds.Height) * scale)
	return clampRect(image.Rect(x1, y1, x2, y2), imgBounds)
}

func clampRect(rect, bounds image.Rectangle) image.Rectangle {
	x1, y1, x2, y2 := rect.Min.X, rect.Min.Y, rect.Max.X, rect.Max.Y
	if x1 < bounds.Min.X {
		x1 = bounds.Min.X
	}
	if y1 < bounds.Min.Y {
		y1 = bounds.Min.Y
	}
	if x2 > bounds.Max.X {
		x2 = bounds.Max.X
	}
	if y2 > bounds.Max.Y {
		y2 = bounds.Max.Y
	}
	return image.Rect(x1, y1, x2, y2)
}

func cropImage(img image.Image, rect image.Rectangle) image.Image {
	return img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(rect)
}

func findWindowByID(ctx context.Context, windowID uint32) (*Window, error) {
	windows, err := ListWindows(ctx)
	if err != nil {
		return nil, fmt.Errorf("list windows: %w", err)
	}

	for i := range windows {
		if windows[i].WindowID == windowID {
			return &windows[i], nil
		}
	}
	return nil, fmt.Errorf("window %d not found", windowID)
}

// MouseMove moves the mouse cursor to the specified coordinates.
// x, y are pixel coordinates in the screenshot image.
func MouseMove(ctx context.Context, windowID uint32, x, y float64) error {
	_, _, xPt, yPt, err := mapWindowInputPoint(ctx, windowID, x, y)
	if err != nil {
		return err
	}

	C.post_mouse_move(C.double(xPt), C.double(yPt))
	return nil
}

// MouseDown sends a mouse down event at the specified coordinates.
func MouseDown(ctx context.Context, windowID uint32, x, y float64, button string) error {
	return postMouseButton(ctx, windowID, x, y, button, func(xPt, yPt C.double, btn C.int) {
		C.post_mouse_down(xPt, yPt, btn)
	})
}

// MouseUp sends a mouse up event at the specified coordinates.
func MouseUp(ctx context.Context, windowID uint32, x, y float64, button string) error {
	return postMouseButton(ctx, windowID, x, y, button, func(xPt, yPt C.double, btn C.int) {
		C.post_mouse_up(xPt, yPt, btn)
	})
}

func postMouseButton(
	ctx context.Context,
	windowID uint32,
	x, y float64,
	button string,
	send func(C.double, C.double, C.int),
) error {
	_, _, xPt, yPt, err := mapWindowInputPoint(ctx, windowID, x, y)
	if err != nil {
		return err
	}
	btn := buttonToInt(button)
	send(C.double(xPt), C.double(yPt), C.int(btn))
	return nil
}

func mapWindowInputPoint(ctx context.Context, windowID uint32, x, y float64) (*Window, *ScreenshotMetadata, float64, float64, error) {
	windows, err := ListWindows(ctx)
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("list windows: %w", err)
	}

	var targetWindow *Window
	for i := range windows {
		if windows[i].WindowID == windowID {
			targetWindow = &windows[i]
			break
		}
	}

	if targetWindow == nil {
		return nil, nil, 0, 0, fmt.Errorf("window %d not found", windowID)
	}

	_, metadata, err := TakeWindowScreenshot(ctx, windowID, imgencode.Options{Quality: 1})
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("take screenshot for coordinate mapping: %w", err)
	}

	x = clampCoord(x, float64(metadata.ImageWidth))
	y = clampCoord(y, float64(metadata.ImageHeight))

	xPt := targetWindow.Bounds.X + (x / metadata.Scale)
	yPt := targetWindow.Bounds.Y + (y / metadata.Scale)

	return targetWindow, metadata, xPt, yPt, nil
}

func mapWindowDragPoints(
	ctx context.Context,
	windowID uint32,
	fromX, fromY, toX, toY float64,
) (*Window, *ScreenshotMetadata, float64, float64, float64, float64, error) {
	targetWindow, metadata, fromXPt, fromYPt, err := mapWindowInputPoint(ctx, windowID, fromX, fromY)
	if err != nil {
		return nil, nil, 0, 0, 0, 0, err
	}

	toX = clampCoord(toX, float64(metadata.ImageWidth))
	toY = clampCoord(toY, float64(metadata.ImageHeight))

	toXPt := targetWindow.Bounds.X + (toX / metadata.Scale)
	toYPt := targetWindow.Bounds.Y + (toY / metadata.Scale)

	return targetWindow, metadata, fromXPt, fromYPt, toXPt, toYPt, nil
}

// Drag performs a drag operation from one point to another.
func Drag(ctx context.Context, windowID uint32, fromX, fromY, toX, toY float64, button string) error {
	_, _, fromXPt, fromYPt, toXPt, toYPt, err := mapWindowDragPoints(ctx, windowID, fromX, fromY, toX, toY)
	if err != nil {
		return err
	}
	btn := buttonToInt(button)
	C.post_mouse_down(C.double(fromXPt), C.double(fromYPt), C.int(btn))
	C.post_mouse_move(C.double(toXPt), C.double(toYPt))
	C.post_mouse_up(C.double(toXPt), C.double(toYPt), C.int(btn))
	return nil
}

// Scroll performs a scroll operation at the specified coordinates.
// deltaX and deltaY are in pixels (positive = right/down).
func Scroll(ctx context.Context, windowID uint32, x, y, deltaX, deltaY float64) error {
	_, _, xPt, yPt, err := mapWindowInputPoint(ctx, windowID, x, y)
	if err != nil {
		return err
	}
	C.post_scroll(C.double(xPt), C.double(yPt), C.double(deltaX), C.double(deltaY))
	return nil
}

func clampCoord(val, maxValue float64) float64 {
	if val < 0 {
		return 0
	}
	if val >= maxValue {
		return maxValue - 1
	}
	return val
}

func buttonToInt(button string) int {
	switch button {
	case "right":
		return 1
	case "middle":
		return 2
	default:
		return 0
	}
}

// CheckPermissions checks if required permissions are granted
func CheckPermissions() (screenRecording bool, accessibility bool) {
	screenRecording = C.has_screen_capture_access() != 0
	accessibility = C.has_accessibility_access() != 0
	return
}

// PermissionError indicates missing macOS permissions required by the tool.
type PermissionError struct {
	ToolName      string
	Screen        bool
	Accessibility bool
}

func (e *PermissionError) Error() string {
	if e.ToolName == "" {
		return "required macOS permissions are not granted"
	}
	return fmt.Sprintf("%s requires macOS permissions: screen recording=%t accessibility=%t. enable both in System Settings > Privacy & Security", e.ToolName, e.Screen, e.Accessibility)
}

// EnsureAutomationPermissions returns an explicit error when screen recording/accessibility are missing.
func EnsureAutomationPermissions(toolName string) error {
	screenRecording, accessibility := CheckPermissions()
	if screenRecording && accessibility {
		return nil
	}
	return &PermissionError{
		ToolName:      toolName,
		Screen:        screenRecording,
		Accessibility: accessibility,
	}
}
