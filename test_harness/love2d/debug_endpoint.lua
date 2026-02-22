-- Debug endpoint for LÖVE Test Harness
-- Provides UDP socket for external control and state querying

local DebugEndpoint = {
    socket = nil,
    port = 9876,
    isEnabled = true,
    messageLog = {},
    maxLogSize = 100,
}

function DebugEndpoint.load()
    if not DebugEndpoint.isEnabled then
        print("Debug endpoint disabled")
        return
    end
    
    local socket = require("socket")
    DebugEndpoint.socket = socket.udp()
    DebugEndpoint.socket:settimeout(0)  -- Non-blocking
    
    local success, err = DebugEndpoint.socket:bind("127.0.0.1", DebugEndpoint.port)
    if success then
        print(string.format("Debug endpoint listening on UDP port %d", DebugEndpoint.port))
        print("Commands: GET_STATE, RESET, PAUSE, RESUME, QUIT, SCREENSHOT")
    else
        print(string.format("Failed to bind debug endpoint: %s", err))
        DebugEndpoint.socket = nil
    end
end

function DebugEndpoint.update(dt)
    if not DebugEndpoint.socket then return end
    
    while true do
        local data, ip, port = DebugEndpoint.socket:receivefrom()
        if not data then break end
        
        local response = DebugEndpoint.handleCommand(data)
        DebugEndpoint.socket:sendto(response, ip, port)
    end
end

function DebugEndpoint.handleCommand(cmd)
    cmd = cmd:upper():gsub("%s+", " "):gsub("^%s*", ""):gsub("%s*$", "")
    
    DebugEndpoint.logMessage(string.format("RECV: %s", cmd))
    
    if cmd == "GET_STATE" then
        return DebugEndpoint.getStateJson()
    elseif cmd == "RESET" then
        -- Trigger reset via global love.load
        if _G.TestHarness then
            _G.TestHarness:resetState()
            return '{"status": "ok", "action": "reset"}'
        else
            return '{"status": "error", "message": "TestHarness not available"}'
        end
    elseif cmd == "PAUSE" then
        if _G.TestHarness then
            _G.TestHarness.state.isPaused = true
            _G.TestHarness:updateStatusLabel()
            return '{"status": "ok", "action": "pause"}'
        else
            return '{"status": "error", "message": "TestHarness not available"}'
        end
    elseif cmd == "RESUME" then
        if _G.TestHarness then
            _G.TestHarness.state.isPaused = false
            _G.TestHarness:updateStatusLabel()
            return '{"status": "ok", "action": "resume"}'
        else
            return '{"status": "error", "message": "TestHarness not available"}'
        end
    elseif cmd == "QUIT" then
        love.event.quit()
        return '{"status": "ok", "action": "quit"}'
    elseif cmd == "SCREENSHOT" then
        -- Return screenshot path if we implement saving
        return string.format('{"status": "ok", "action": "screenshot", "note": "Use love.graphics.captureScreenshot() in LÖVE 11.0+"}')
    elseif cmd:sub(1, 9) == "SET_SCENE" then
        local scene = cmd:sub(11):gsub("%s", "")
        if _G.TestHarness and scene ~= "" then
            _G.TestHarness.state.currentScene = scene
            _G.TestHarness:updateStatusLabel()
            return string.format('{"status": "ok", "action": "set_scene", "scene": "%s"}', scene)
        else
            return '{"status": "error", "message": "Invalid scene or TestHarness not available"}'
        end
    elseif cmd == "PING" then
        return '{"status": "ok", "pong": true}'
    else
        return string.format('{"status": "error", "message": "Unknown command: %s", "available": ["GET_STATE", "RESET", "PAUSE", "RESUME", "QUIT", "SCREENSHOT", "SET_SCENE", "PING"]}', cmd)
    end
end

function DebugEndpoint.getStateJson()
    local state = love.getTestState and love.getTestState() or {}
    
    -- Build JSON manually for simplicity
    local json = string.format(
        '{"status": "ok", "state": {"isPaused": %s, "currentScene": "%s", "buttonClicks": {"start": %d, "reset": %d, "pause": %d}, "windowInfo": {"title": "%s", "width": %d, "height": %d}}}',
        tostring(state.isPaused or false),
        state.currentScene or "unknown",
        (state.buttonClicks and state.buttonClicks.start) or 0,
        (state.buttonClicks and state.buttonClicks.reset) or 0,
        (state.buttonClicks and state.buttonClicks.pause) or 0,
        (state.windowInfo and state.windowInfo.title) or "unknown",
        (state.windowInfo and state.windowInfo.width) or 0,
        (state.windowInfo and state.windowInfo.height) or 0
    )
    
    return json
end

function DebugEndpoint.logMessage(msg)
    table.insert(DebugEndpoint.messageLog, {
        time = os.time(),
        message = msg,
    })
    
    -- Trim log if too large
    while #DebugEndpoint.messageLog > DebugEndpoint.maxLogSize do
        table.remove(DebugEndpoint.messageLog, 1)
    end
end

function DebugEndpoint.draw()
    -- Optional: draw debug overlay
    if not DebugEndpoint.isEnabled then return end
    
    love.graphics.setColor(0, 0, 0, 0.5)
    love.graphics.rectangle("fill", 10, love.graphics.getHeight() - 100, 300, 90)
    
    love.graphics.setColor(0, 1, 0)
    love.graphics.setFont(love.graphics.newFont(10))
    love.graphics.print(string.format("Debug: UDP port %d", DebugEndpoint.port), 15, love.graphics.getHeight() - 95)
    
    -- Show last few messages
    local y = love.graphics.getHeight() - 80
    local count = 0
    for i = #DebugEndpoint.messageLog, 1, -1 do
        if count >= 3 then break end
        local entry = DebugEndpoint.messageLog[i]
        love.graphics.print(string.format("[%d] %s", entry.time % 1000, entry.message:sub(1, 40)), 15, y)
        y = y + 15
        count = count + 1
    end
end

return DebugEndpoint
