package window

/*
#cgo LDFLAGS: -framework CoreGraphics -framework CoreFoundation -framework ApplicationServices
#include <CoreGraphics/CoreGraphics.h>
#include <CoreFoundation/CoreFoundation.h>
#include <ApplicationServices/ApplicationServices.h>
#include <stdlib.h>
#include <string.h>

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
	if windowList == nil {
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
	if dict == nil {
		return nil
	}

	win := &Window{}

	// Get window ID (kCGWindowNumber)
	if num := C.cf_dict_get_number(dict, "kCGWindowNumber"); num != nil {
		win.WindowID = uint32(C.cfnumber_to_int32(num))
	}

	// Get owner name (kCGWindowOwnerName)
	if str := C.cf_dict_get_string(dict, "kCGWindowOwnerName"); str != nil {
		cstr := C.cfstring_to_cstring(str)
		if cstr != nil {
			win.OwnerName = C.GoString(cstr)
			C.free(unsafe.Pointer(cstr))
		}
	}

	// Get PID (kCGWindowOwnerPID)
	if num := C.cf_dict_get_number(dict, "kCGWindowOwnerPID"); num != nil {
		win.PID = C.cfnumber_to_int32(num)
	}

	// Get title (kCGWindowName)
	if str := C.cf_dict_get_string(dict, "kCGWindowName"); str != nil {
		cstr := C.cfstring_to_cstring(str)
		if cstr != nil {
			win.Title = C.GoString(cstr)
			C.free(unsafe.Pointer(cstr))
		}
	}

	// Get bounds (kCGWindowBounds)
	if boundsDict := C.cf_dict_get_dict(dict, "kCGWindowBounds"); boundsDict != nil {
		win.Bounds = parseBounds(boundsDict)
	}

	// Check if on screen (kCGWindowIsOnscreen)
	if num := C.cf_dict_get_number(dict, "kCGWindowIsOnscreen"); num != nil {
		win.IsOnScreen = C.cfnumber_to_int32(num) != 0
	} else {
		win.IsOnScreen = true // Default to true if not present
	}

	return win
}

func parseBounds(dict C.CFDictionaryRef) Bounds {
	b := Bounds{}

	if x := C.cf_dict_get_number(dict, "X"); x != nil {
		b.X = C.cfnumber_to_double(x)
	}
	if y := C.cf_dict_get_number(dict, "Y"); y != nil {
		b.Y = C.cfnumber_to_double(y)
	}
	if w := C.cf_dict_get_number(dict, "Width"); w != nil {
		b.Width = C.cfnumber_to_double(w)
	}
	if h := C.cf_dict_get_number(dict, "Height"); h != nil {
		b.Height = C.cfnumber_to_double(h)
	}

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

	// Capture the window
	cgWindowID := C.CGWindowID(windowID)

	// Create bounds CGRect
	bounds := C.CGRectMake(
		C.CGFloat(targetWindow.Bounds.X),
		C.CGFloat(targetWindow.Bounds.Y),
		C.CGFloat(targetWindow.Bounds.Width),
		C.CGFloat(targetWindow.Bounds.Height),
	)

	// Capture window
	imageRef := C.CGWindowListCreateImage(
		bounds,
		C.kCGWindowListOptionIncludingWindow,
		cgWindowID,
		C.kCGWindowImageBoundsIgnoreFraming|C.kCGWindowImageShouldBeOpaque,
	)

	if imageRef == nil {
		return nil, nil, fmt.Errorf("failed to capture window %d: screen recording permission may be required", windowID)
	}
	defer C.CGImageRelease(imageRef)

	// Get image dimensions
	width := int(C.CGImageGetWidth(imageRef))
	height := int(C.CGImageGetHeight(imageRef))

	// Calculate scale (assuming Retina displays have 2x scale)
	scale := float64(width) / targetWindow.Bounds.Width

	// Convert to Go image
	img := cgImageToGoImage(imageRef)
	if img == nil {
		return nil, nil, fmt.Errorf("failed to convert captured image")
	}

	// Encode to JPEG
	data, err := imgencode.EncodeJPEG(img, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("encode screenshot: %w", err)
	}

	metadata := &ScreenshotMetadata{
		WindowID:    windowID,
		Bounds:      targetWindow.Bounds,
		ImageWidth:  width,
		ImageHeight: height,
		Scale:       scale,
	}

	return data, metadata, nil
}

func cgImageToGoImage(cgImage C.CGImageRef) image.Image {
	// Get image properties
	width := int(C.CGImageGetWidth(cgImage))
	height := int(C.CGImageGetHeight(cgImage))

	// Create bitmap context to draw into
	colorSpace := C.CGColorSpaceCreateDeviceRGB()
	defer C.CGColorSpaceRelease(colorSpace)

	bitsPerComponent := 8
	bitsPerPixel := 32
	bytesPerRow := width * 4

	// Allocate buffer
	bufferSize := bytesPerRow * height
	buffer := make([]byte, bufferSize)

	context := C.CGBitmapContextCreate(
		unsafe.Pointer(&buffer[0]),
		C.size_t(width),
		C.size_t(height),
		C.size_t(bitsPerComponent),
		C.size_t(bytesPerRow),
		colorSpace,
		C.kCGImageAlphaPremultipliedLast,
	)
	if context == nil {
		return nil
	}
	defer C.CGContextRelease(context)

	// Draw the image
	rect := C.CGRectMake(0, 0, C.CGFloat(width), C.CGFloat(height))
	C.CGContextDrawImage(context, rect, cgImage)

	// Create RGBA image
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			offset := y*bytesPerRow + x*4
			rgba.Set(x, y, image.RGBA{
				R: buffer[offset],
				G: buffer[offset+1],
				B: buffer[offset+2],
				A: buffer[offset+3],
			})
		}
	}

	return rgba
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
