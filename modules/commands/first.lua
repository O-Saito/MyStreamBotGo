function on_command(ev)
    if #ev.Args > 0 and ev.Args[1] ~= "" then
        local target = ev.Args[1]
        if string.find(target, "@") then
            target = string.gsub(target, "@", "")
        end
        local message = "O " .. target .. " nunca foi o primeirinho!"
        local u = find_user_data(ev.Source, target)
        if u ~= nil then
            message = "O " .. u.display .. " já foi o primeiro " .. u.count .. " vez(es)!"
        end
        g.send_message(ev.Source, ev.Channel, message, ev.Message.MessageId)
        return
    end

    local first = g.get('first_user_' .. ev.Source)
    if first == nil then
        local user = ev.Message.User
        local display = ev.Message.User
        if ev.Message.Metadata['display-name'] ~= nil then
            display = ev.Message.Metadata['display-name']
        end
        g.set('first_user_' .. ev.Source, display)

        local var_name = 'first_' .. ev.Source .. '_user_count'
        local f = g.kv_get(var_name)
        if f == nil then
            f = {}
        end

        local userId = ev.Message.UserId
        if f[userId] == nil then
            f[userId] = {
                user = display,
                login = user,
                count = 0
            }
        end

        local d = f[userId]

        d.user = user
        d.display = display
        d.count = d.count + 1

        g.kv_set(var_name, f)

        g.send_message(ev.Source, ev.Channel,
            "O primeirinho foi você! E já foi o primeirinho " .. d.count .. " vez(es)!", ev.Message.MessageId)
        return
    end

    local message = "O primeirinho foi o " .. first
    local u = find_user_data(ev.Source, ev.Message.User)

    if u ~= nil then
        message = message .. " ele já foi o primeirinho " .. u.count .. " vez(es)!"
    end

    g.send_message(ev.Source, ev.Channel, message, ev.Message.MessageId)
end

function find_user_data(source, username)
    local var_name = 'first_' .. source .. '_user_count'
    local f = g.kv_get(var_name)
    local u = nil

    if f == nil then
        return nil
    end

    for k, v in pairs(f) do
        if v.login == username then
            u = v
            break
        end
    end
    return u
end
