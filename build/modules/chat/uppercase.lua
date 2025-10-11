function on_message(ev)
    local user = ev.User
    local text = ev.Message
    local upper = string.upper(text)
    main.print("[UPPERCASE] " .. user .. ": " .. upper)
end