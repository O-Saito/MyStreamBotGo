local ignore = {
    ['streamelements'] = true,
    ['streamlabs'] = true,
}

function on_message(ev)

    if ev.User == nil or ignore[ev.User] then
        return
    end

    local user = ev.User
    if g.get('first_user_'..ev.Source) ~= nil then
        return
    end
    local display = ev.User
    if ev.Metadata['display-name'] ~= nil then
        display = ev.Metadata['display-name']
    end
    g.set('first_user_'..ev.Source, display)
    
    local var_name = 'first_'..ev.Source..'_user_count'
    local f = g.kv_get(var_name)
    if f == nil then
        f = {}
    end

    if f[ev.UserId] == nil then
        f[ev.UserId] = { user = display, login = user, count = 0 }
    end

    f[ev.UserId].count = f[ev.UserId].count + 1
    f[ev.UserId].user = user
    f[ev.UserId].display = display
    
    g.kv_set(var_name, f)
end