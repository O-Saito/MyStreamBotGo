function on_start()
    ev.set_interval(5)
    ev.set_paused(false)
    ev.data = ev.data or { count = 0 }
    ev.use_db()
    ev.db_exec("CREATE TABLE IF NOT EXISTS logs (id INTEGER PRIMARY KEY, msg TEXT)")
end

function on_tick()
    ev.data.count = ev.data.count + 1
    ev.socket_send("tick", { count = ev.data.count })
end

function on_event(name, data)
    g.log("Event: " .. name, data)
end

function on_message(msg)
    g.log("Chat from " .. msg.User .. ": " .. msg.Message)
end

function on_command(name, data)
    g.log("Command !" .. name, data)
end

function on_request(type, data)
    if type == "ping" then
        data.respond({ pong = true, count = ev.data.count })
    elseif type == "logs" then
        local rows = ev.db_query("SELECT * FROM logs")
        data.respond({ logs = rows })
    end
end
