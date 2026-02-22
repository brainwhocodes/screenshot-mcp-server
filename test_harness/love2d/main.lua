-- LÖVE Test Harness for UI Automation Testing
-- This game provides a safe, deterministic target for testing the screenshot-mcp-server

local DebugEndpoint = require("debug_endpoint")

-- Make TestHarness global so debug endpoint can access it
_G.TestHarness = {
    -- Configuration
    -- Configuration
    config = {
        windowTitle = "UI Automation Test Target",
        windowWidth = 800,
        windowHeight = 600,
        highdpi = false,  -- Use false for simpler coordinate mapping
        fixedTimestep = true,
        seed = 12345,  -- Fixed RNG seed for determinism
    },
    
    -- State
    state = {
        isPaused = true,  -- Start paused for safety
        currentScene = "menu",
        buttonClicks = {},
        lastClickTime = 0,
        fps = 0,
    },
    
    -- UI Elements
    ui = {
        buttons = {},
        labels = {},
    },
}

function love.conf(t)
    t.window.title = TestHarness.config.windowTitle
    t.window.width = TestHarness.config.windowWidth
    t.window.height = TestHarness.config.windowHeight
    t.window.resizable = false
    t.window.fullscreen = false
    t.window.highdpi = TestHarness.config.highdpi
    t.window.vsync = 1
    
    t.modules.joystick = false
    t.modules.physics = false
    t.modules.thread = false
end

function love.load()
    -- Set fixed RNG seed for deterministic behavior
    math.randomseed(TestHarness.config.seed)
    
    -- Set window title programmatically (ensures it's correct)
    love.window.setTitle(TestHarness.config.windowTitle)
    
    -- Initialize fonts
    TestHarness.fonts = {
        title = love.graphics.newFont(32),
        normal = love.graphics.newFont(16),
        small = love.graphics.newFont(12),
    }
    
    -- Setup UI
    TestHarness:setupUI()
    
    -- Initialize debug endpoint
    DebugEndpoint.load()
    
    print("Test Harness loaded - Window is in SAFE MODE (paused)")
    print("Click anywhere or press SPACE to start interacting")
end

function TestHarness:setupUI()
    local w, h = love.graphics.getDimensions()
    local centerX = w / 2
    local buttonWidth = 200
    local buttonHeight = 50
    local startY = h / 2 - 100
    
    -- Main menu buttons
    self.ui.buttons = {
        {
            id = "start",
            text = "Start Test",
            x = centerX - buttonWidth/2,
            y = startY,
            width = buttonWidth,
            height = buttonHeight,
            color = {0.2, 0.7, 0.2},
            hoverColor = {0.3, 0.8, 0.3},
            clicked = false,
            clickCount = 0,
        },
        {
            id = "reset",
            text = "Reset State",
            x = centerX - buttonWidth/2,
            y = startY + 70,
            width = buttonWidth,
            height = buttonHeight,
            color = {0.7, 0.5, 0.2},
            hoverColor = {0.8, 0.6, 0.3},
            clicked = false,
            clickCount = 0,
        },
        {
            id = "pause",
            text = "Pause/Resume",
            x = centerX - buttonWidth/2,
            y = startY + 140,
            width = buttonWidth,
            height = buttonHeight,
            color = {0.2, 0.5, 0.7},
            hoverColor = {0.3, 0.6, 0.8},
            clicked = false,
            clickCount = 0,
        },
    }
    
    -- Status labels
    self.ui.labels = {
        {
            text = "Status: PAUSED (Safe Mode)",
            x = 20,
            y = 20,
            color = {1, 0.5, 0.5},
        },
        {
            text = "Window: " .. tostring(w) .. "x" .. tostring(h),
            x = 20,
            y = 50,
            color = {0.8, 0.8, 0.8},
        },
        {
            text = "Scene: menu",
            x = 20,
            y = 70,
            color = {0.8, 0.8, 0.8},
        },
    }
end

function love.update(dt)
    if TestHarness.config.fixedTimestep then
        -- Fixed timestep for determinism
        dt = 1/60
    end
    
    TestHarness.state.fps = love.timer.getFPS()
    
    if not TestHarness.state.isPaused then
        -- Update game state only when not paused
        TestHarness:updateGame(dt)
    end
    
    -- Update button hover states
    local mx, my = love.mouse.getPosition()
    for _, button in ipairs(TestHarness.ui.buttons) do
        button.isHovered = mx >= button.x and mx <= button.x + button.width and
                           my >= button.y and my <= button.y + button.height
    end
    
    -- Update debug endpoint
    DebugEndpoint.update(dt)
end

function TestHarness:updateGame(dt)
    -- Minimal game update - just track time
    self.state.lastClickTime = self.state.lastClickTime + dt
end

function love.draw()
    -- Clear background
    love.graphics.setBackgroundColor(0.15, 0.15, 0.2)
    
    -- Draw title
    love.graphics.setFont(TestHarness.fonts.title)
    love.graphics.setColor(1, 1, 1)
    love.graphics.printf("UI Automation Test Target", 0, 80, love.graphics.getWidth(), "center")
    
    -- Draw buttons
    love.graphics.setFont(TestHarness.fonts.normal)
    for _, button in ipairs(TestHarness.ui.buttons) do
        -- Button background
        if button.isHovered then
            love.graphics.setColor(button.hoverColor)
        else
            love.graphics.setColor(button.color)
        end
        love.graphics.rectangle("fill", button.x, button.y, button.width, button.height, 5, 5)
        
        -- Button border (highlight if clicked)
        if button.clicked then
            love.graphics.setColor(1, 1, 0)
            love.graphics.setLineWidth(3)
        else
            love.graphics.setColor(0.3, 0.3, 0.3)
            love.graphics.setLineWidth(1)
        end
        love.graphics.rectangle("line", button.x, button.y, button.width, button.height, 5, 5)
        
        -- Button text
        love.graphics.setColor(1, 1, 1)
        love.graphics.printf(button.text, button.x, button.y + button.height/2 - 10, button.width, "center")
        
        -- Click counter
        love.graphics.setFont(TestHarness.fonts.small)
        love.graphics.setColor(0.8, 0.8, 0.8)
        love.graphics.print("Clicks: " .. button.clickCount, button.x + button.width + 10, button.y + button.height/2 - 5)
        love.graphics.setFont(TestHarness.fonts.normal)
    end
    
    -- Draw labels
    love.graphics.setFont(TestHarness.fonts.small)
    for i, label in ipairs(TestHarness.ui.labels) do
        love.graphics.setColor(label.color)
        love.graphics.print(label.text, label.x, label.y)
    end
    
    -- Draw instructions
    love.graphics.setColor(0.6, 0.6, 0.6)
    love.graphics.printf(
        "This window is safe for automated testing.\n" ..
        "All clicks are contained within this window.\n" ..
        "Press ESC to toggle pause mode.",
        0, love.graphics.getHeight() - 80, love.graphics.getWidth(), "center"
    )
    
    -- Draw FPS
    love.graphics.setColor(0.5, 0.5, 0.5)
    love.graphics.print("FPS: " .. TestHarness.state.fps, love.graphics.getWidth() - 80, 20)
    
    -- Draw debug overlay
    DebugEndpoint.draw()
end

function love.mousepressed(x, y, button)
    if button ~= 1 then return end  -- Only handle left clicks
    
    TestHarness.state.lastClickTime = 0
    
    -- Check button clicks
    for _, btn in ipairs(TestHarness.ui.buttons) do
        if x >= btn.x and x <= btn.x + btn.width and
           y >= btn.y and y <= btn.y + btn.height then
            TestHarness:handleButtonClick(btn)
            return
        end
    end
    
    -- Clicking anywhere else resumes from pause
    if TestHarness.state.isPaused then
        TestHarness.state.isPaused = false
        TestHarness:updateStatusLabel()
        print("Resumed from pause")
    end
end

function TestHarness:handleButtonClick(button)
    button.clicked = true
    button.clickCount = button.clickCount + 1
    
    -- Reset clicked state after a short delay
    love.timer.sleep(0.1)
    button.clicked = false
    
    if button.id == "start" then
        print("Start Test button clicked")
        self.state.currentScene = "test"
        self.state.isPaused = false
    elseif button.id == "reset" then
        print("Reset State button clicked")
        self:resetState()
    elseif button.id == "pause" then
        self.state.isPaused = not self.state.isPaused
        print("Pause toggled:", self.state.isPaused)
    end
    
    self:updateStatusLabel()
end

function TestHarness:resetState()
    -- Reset to deterministic initial state
    math.randomseed(self.config.seed)
    self.state.currentScene = "menu"
    self.state.isPaused = true
    self.state.lastClickTime = 0
    
    for _, button in ipairs(self.ui.buttons) do
        button.clickCount = 0
        button.clicked = false
    end
    
    print("State reset to initial values")
    self:updateStatusLabel()
end

function TestHarness:updateStatusLabel()
    local statusText = "Status: "
    if self.state.isPaused then
        statusText = statusText .. "PAUSED (Safe Mode)"
        self.ui.labels[1].color = {1, 0.5, 0.5}
    else
        statusText = statusText .. "ACTIVE"
        self.ui.labels[1].color = {0.5, 1, 0.5}
    end
    self.ui.labels[1].text = statusText
    self.ui.labels[3].text = "Scene: " .. self.state.currentScene
end

function love.keypressed(key)
    if key == "escape" then
        TestHarness.state.isPaused = not TestHarness.state.isPaused
        TestHarness:updateStatusLabel()
        print("Pause toggled via ESC:", TestHarness.state.isPaused)
    elseif key == "r" and love.keyboard.isDown("lctrl", "rctrl") then
        TestHarness:resetState()
    elseif key == "q" and love.keyboard.isDown("lctrl", "rctrl") then
        love.event.quit()
    end
end

function love.focus(focused)
    if not focused then
        -- Auto-pause when window loses focus for safety
        TestHarness.state.isPaused = true
        TestHarness:updateStatusLabel()
        print("Window lost focus - auto-paused")
    end
end

-- Expose state for external debugging/assertions
function love.getTestState()
    return {
        isPaused = TestHarness.state.isPaused,
        currentScene = TestHarness.state.currentScene,
        buttonClicks = {
            start = TestHarness.ui.buttons[1].clickCount,
            reset = TestHarness.ui.buttons[2].clickCount,
            pause = TestHarness.ui.buttons[3].clickCount,
        },
        windowInfo = {
            title = TestHarness.config.windowTitle,
            width = TestHarness.config.windowWidth,
            height = TestHarness.config.windowHeight,
        },
    }
end
