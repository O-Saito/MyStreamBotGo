function on_message(ev)
    local user = ev.User
    local text = ev.Message
    local upper = string.upper(text)
    g.print("[UPPERCASE] " .. user .. ": " .. upper)
end