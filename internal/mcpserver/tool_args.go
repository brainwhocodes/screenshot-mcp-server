package mcpserver

type emptyArgs struct{}

type mouseButtonArgs struct {
	WindowID uint32  `json:"window_id"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Button   string  `json:"button,omitempty"`
}

type screenshotArgs struct{}

type takeScreenshotPNGArgs struct{}

type listWindowsArgs struct{}

type screenshotHashArgs struct {
	Algorithm     string `json:"algorithm,omitempty"`
	Target        string `json:"target,omitempty"`
	WindowID      uint32 `json:"window_id,omitempty"`
	IncludeCursor bool   `json:"include_cursor,omitempty"`
}

type focusWindowArgs struct {
	WindowID uint32 `json:"window_id"`
}

type takeWindowScreenshotArgs struct {
	WindowID uint32 `json:"window_id"`
}

type takeWindowScreenshotPNGArgs struct {
	WindowID uint32 `json:"window_id"`
}

type takeRegionScreenshotArgs struct {
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Width      float64 `json:"width"`
	Height     float64 `json:"height"`
	CoordSpace string  `json:"coord_space,omitempty"`
}

type takeRegionScreenshotPNGArgs struct {
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Width      float64 `json:"width"`
	Height     float64 `json:"height"`
	CoordSpace string  `json:"coord_space,omitempty"`
}

type clickArgs struct {
	WindowID uint32  `json:"window_id"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Button   string  `json:"button,omitempty"`
	Clicks   int     `json:"clicks,omitempty"`
}

type dragArgs struct {
	WindowID uint32  `json:"window_id"`
	FromX    float64 `json:"from_x"`
	FromY    float64 `json:"from_y"`
	ToX      float64 `json:"to_x"`
	ToY      float64 `json:"to_y"`
	Button   string  `json:"button,omitempty"`
}

type scrollArgs struct {
	WindowID uint32  `json:"window_id"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	DeltaX   float64 `json:"delta_x"`
	DeltaY   float64 `json:"delta_y"`
}

type pressKeyArgs struct {
	WindowID   uint32   `json:"window_id"`
	Key        string   `json:"key"`
	Modifiers  []string `json:"modifiers,omitempty"`
	KeyPresses int      `json:"presses,omitempty"`
}

type typeTextArgs struct {
	WindowID uint32 `json:"window_id"`
	Text     string `json:"text"`
	DelayMs  int    `json:"delay_ms,omitempty"`
}

type keyActionArgs struct {
	WindowID  uint32   `json:"window_id"`
	Key       string   `json:"key"`
	Modifiers []string `json:"modifiers,omitempty"`
}

type waitForPixelArgs struct {
	WindowID       uint32   `json:"window_id"`
	X              float64  `json:"x"`
	Y              float64  `json:"y"`
	RGBA           [4]uint8 `json:"rgba"`
	Tolerance      int      `json:"tolerance,omitempty"`
	TimeoutMs      int      `json:"timeout_ms,omitempty"`
	PollIntervalMs int      `json:"poll_interval_ms,omitempty"`
}

type waitForRegionStableArgs struct {
	WindowID       uint32  `json:"window_id"`
	X              float64 `json:"x"`
	Y              float64 `json:"y"`
	Width          float64 `json:"width"`
	Height         float64 `json:"height"`
	StableCount    int     `json:"stable_count,omitempty"`
	TimeoutMs      int     `json:"timeout_ms,omitempty"`
	PollIntervalMs int     `json:"poll_interval_ms,omitempty"`
}

type launchAppArgs struct {
	AppName string `json:"app_name"`
}

type quitAppArgs struct {
	AppName string `json:"app_name"`
}

type waitForProcessArgs struct {
	ProcessName    string `json:"process_name"`
	TimeoutMs      int    `json:"timeout_ms,omitempty"`
	PollIntervalMs int    `json:"poll_interval_ms,omitempty"`
}

type killProcessArgs struct {
	ProcessName string `json:"process_name"`
}

type setClipboardArgs struct {
	Text string `json:"text"`
}

type waitForImageMatchArgs struct {
	WindowID       uint32  `json:"window_id,omitempty"`
	TemplateImage  string  `json:"template_image"`
	Threshold      float64 `json:"threshold,omitempty"`
	TimeoutMs      int     `json:"timeout_ms,omitempty"`
	PollIntervalMs int     `json:"poll_interval_ms,omitempty"`
}

type findImageMatchesArgs struct {
	WindowID      uint32  `json:"window_id,omitempty"`
	TemplateImage string  `json:"template_image"`
	Threshold     float64 `json:"threshold,omitempty"`
}

type compareImagesArgs struct {
	Image1    string  `json:"image1"`
	Image2    string  `json:"image2"`
	Threshold float64 `json:"threshold,omitempty"`
}

type assertScreenshotMatchesFixtureArgs struct {
	WindowID    uint32       `json:"window_id"`
	FixturePath string       `json:"fixture_path"`
	Threshold   float64      `json:"threshold,omitempty"`
	MaskRegions []MaskRegion `json:"mask_regions,omitempty"`
}

type waitForTextArgs struct {
	WindowID       uint32 `json:"window_id,omitempty"`
	Text           string `json:"text"`
	TimeoutMs      int    `json:"timeout_ms,omitempty"`
	PollIntervalMs int    `json:"poll_interval_ms,omitempty"`
}

type restartAppArgs struct {
	AppName string `json:"app_name"`
}

type startRecordingArgs struct {
	WindowID uint32 `json:"window_id,omitempty"`
	FPS      int    `json:"fps,omitempty"`
	Format   string `json:"format,omitempty"`
}

type stopRecordingArgs struct {
	RecordingID string `json:"recording_id"`
}
