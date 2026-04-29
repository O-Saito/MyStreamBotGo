function on_start()
    ev.set_interval(0)
    ev.set_paused(false)
end

function on_request(name, data)
    g.print(name)
    if name == 'get_functions' then
        local functions = {}
        for k, v in pairs(twitch) do
            table.insert(functions, k)
        end
        data.respond(functions)
        return
    end
    if twitch[name] then
        local args = {}
        for k, v in pairs(data) do
            table.insert(args, k)
        end
        g.print(name, {unpack(data)})
        local d = twitch[name](unpack(data))
        if type(d) == "string" then
            d = { stringargs = d }
        end
        data.respond(d)
    end
end
