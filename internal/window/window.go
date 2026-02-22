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
	"unsafe"

	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/safeexec"
)

// SupportsWindowTools reports whether this platform supports window-level automation features.
func SupportsWindowTools() bool {
	return true
}

func postMouseMoveEvent(x, y float64) {
	C.post_mouse_move(C.double(x), C.double(y))
}

func postMouseDownEvent(x, y float64, button int) {
	C.post_mouse_down(C.double(x), C.double(y), C.int(button))
}

func postMouseUpEvent(x, y float64, button int) {
	C.post_mouse_up(C.double(x), C.double(y), C.int(button))
}

func postMouseClickEvent(x, y float64, button int, clicks int) {
	C.post_mouse_click(C.double(x), C.double(y), C.int(button), C.int(clicks))
}

func postScrollEvent(x, y, deltaX, deltaY float64) {
	C.post_scroll(C.double(x), C.double(y), C.double(deltaX), C.double(deltaY))
}

func scaleAtPoint(x, y float64) float64 {
	return float64(C.get_scale_at_point(C.double(x), C.double(y)))
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
