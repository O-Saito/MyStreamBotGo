function on_message(ev)
    local user = ev.User
    local text = ev.Message
    local upper = string.upper(text)

    --[[local hasViewer = false
    local viewers = state.GetViewers()
    for i, v in ipairs(viewers)  do
        if v == user then
            hasViewer = true
            break
        end
    end

    if not hasViewer then
        state.AddViewer(user)
    end]]

    g.print("[UPPERCASE] " .. user .. ": " .. upper)
    --g.socket_send("chat-message", { type = "message", message = upper, viewers = state:GetViewers() })
end