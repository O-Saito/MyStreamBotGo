local accent_map = {
    ["á"] = "a",
    ["é"] = "e",
    ["í"] = "i",
    ["ó"] = "o",
    ["ú"] = "u",
    ["â"] = "a",
    ["ê"] = "e",
    ["ô"] = "o",
    ["à"] = "a",
    ["ã"] = "a",
    ["õ"] = "o",
    ["Á"] = "A",
    ["É"] = "E",
    ["Í"] = "I",
    ["Ó"] = "O",
    ["Ú"] = "U",
    ["Â"] = "A",
    ["Ê"] = "E",
    ["Ô"] = "O",
    ["À"] = "A",
    ["Ã"] = "A",
    ["Õ"] = "O"
};

options = {"Sim", "Não"}

function higieniza_string(str)
    local lowerStr = string.lower(str)
    for accent, char in pairs(accent_map) do
        lowerStr = string.gsub(lowerStr, accent, char)
    end
    local cleanedStr = string.gsub(lowerStr, '[^%a%d%s]', '')
    return cleanedStr
end

function reset_data()

    ev.data = {
        users_voted = {},
        votos = {},
        alias = {},
        tempo = 20 -- segundos para final da votação
    }

    for k, v in pairs(options) do
        local str = higieniza_string(v)
        ev.data.votos[str] = {
            descr = v,
            mapValue = str,
            indexValue = k,
            count = 0
        }

        table.insert(ev.data.alias, ev.data.votos[str])
    end

end

function on_start()
    print("[Lua] Evento de teste iniciado!")
    ev.setInterval(1)
    ev.setPaused(false)
    reset_data()
    g.socket_send("user_vote_update", {
        votos = ev.data.votos,
        tempo = tempo,
        ended = false
    })
end

function on_tick()
    local tempo = ev.data.tempo
    tempo = tempo - 1
    print("[Lua] Tempo restante:", tempo)
    ev.data.tempo = tempo

    if tempo > 0 and tempo % 10 == 0 then
        g.socket_send("user_vote_update", {
            votos = ev.data.votos,
            tempo = tempo,
            ended = false
        })
    end

    if tempo <= 0 then
        print("[Lua] Votação encerrada!")
        ev.setPaused(true)
        g.socket_send("user_vote_update", {
            votos = ev.data.votos,
            ended = true
        })
        -- g.print(string.format("[VOTAÇÃO] Resultados - Sim: %d, Não: %d", ev.data.votos.sim, ev.data.votos.nao))
        reset_data()
    end
end

function on_message(msg)
    if ev.data.users_voted[msg.UserId] ~= nil then
        return
    end

    local cleanedStr = higieniza_string(msg.Message)

    if ev.data.votos[cleanedStr] == nil then
        local asNumber = tonumber(cleanedStr)
        if asNumber == nil or ev.data.alias[asNumber] == nil then
            g.print("[Lua] Voto inválido de" .. msg.User .. ":" .. msg.Message)
            return
        end
        cleanedStr = ev.data.alias[asNumber].mapValue
    end

    ev.data.votos[cleanedStr].count = ev.data.votos[cleanedStr].count + 1

    ev.data.users_voted[msg.UserId] = true
end

function on_event(msg)
    if msg.payload and msg.payload.type == "vote" then
        local voto = msg.payload.value
        if ev.data.votos[voto] then
            ev.data.votos[voto] = ev.data.votos[voto] + 1
        end
        print("[Lua] Voto recebido:", voto)
    end
end
