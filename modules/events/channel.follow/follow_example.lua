function on_event(data)
    local event = data.payload.event
    g.print("New follower: " .. event.user_name .. " (" .. event.user_id .. ")")
end
