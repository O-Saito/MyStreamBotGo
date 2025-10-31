local string_helper = require("string_helper")

local defaultTempo = 15 -- 60 * 2

function reset_data()
    ev.data = {
        votacao_em_andamento = false,
        users_voted = {},
        votos = {},
        alias = {},
        tempo = defaultTempo
    }
end

function setar_opcoes(options)
    ev.data.votos = {}
    for k, v in pairs(options) do
        local str = string_helper.higieniza_string(v)
        ev.data.votos[str] = {
            descr = v,
            mapValue = str,
            indexValue = k,
            count = 0
        }
        table.insert(ev.data.alias, ev.data.votos[str])
    end
end

function on_request(type, data)
    g.log("[VOTE] Request " .. type, data)
    if type == "setup" then
        setar_opcoes(data.opcoes)
        if data.tempo == nil then
            data.tempo = defaultTempo
        end
        ev.data.tempo = data.tempo
        ev.socket_send('config', ev.data)
    end

    if type == "start" then
        if data.opcoes ~= nil then
            setar_opcoes(data.opcoes)
        end
        if data.tempo ~= nil then
            ev.data.tempo = data.tempo
        end
        ev.data.ended = false
        ev.data.users_voted = {}
        ev.data.votacao_em_andamento = true
    end
end

function on_start()
    print("[VOTE] Evento iniciado!")
    ev.setInterval(1)
    ev.setPaused(false)
    reset_data()
    ev.socket_send("config", {
        votacao_em_andamento = ev.data.votacao_em_andamento,
        votos = ev.data.votos,
        tempo = defaultTempo,
        ended = false
    })
end

function on_tick(data)
    if ev.data.votacao_em_andamento == false then
        return
    end
    local tempo = ev.data.tempo
    tempo = tempo - 1
    ev.data.tempo = tempo

    ev.socket_send("user_vote_update", {
        votos = ev.data.votos,
        tempo = tempo,
        ended = false
    })

    if tempo <= 0 then
        -- ev.setPaused(true)
        ev.socket_send("user_vote_update", {
            tempo = tempo,
            votos = ev.data.votos,
            ended = true
        })
        reset_data()
    end
end

function on_message(msg)
    if ev.data.users_voted[msg.UserId] ~= nil then
        print('Usuário já votou ' .. msg.UserId)
        return
    end

    local cleanedStr = string_helper.higieniza_string(msg.Message)

    if ev.data.votos[cleanedStr] == nil then
        local asNumber = tonumber(cleanedStr)
        -- print(asNumber)
        if asNumber == nil or ev.data.alias[asNumber] == nil then
            print(ev.data.alias)
            return
        end
        cleanedStr = ev.data.alias[asNumber].mapValue
        if ev.data.votos[cleanedStr] == nil then
            print(cleanedStr)
            return
        end
    end

    ev.data.votos[cleanedStr].count = ev.data.votos[cleanedStr].count + 1

    ev.data.users_voted[msg.UserId] = true
end

function on_event(name, data)
    g.print("[VOTE] Evento recebido " .. name, data)
end

function on_command(name, data)
    g.print("[VOTE] Comando recebido:" .. name, data)
    if name == 'vote' and data.Message.Metadata ~= nil and data.Message.Metadata["user-type"] == "mod" then
        local args = string_helper.trim(table.concat(data.Args, " "))
        g.print("[VOTE] ARGS: " .. args)
        local options = string_helper.split(args, ";")
        if #options == 0 then
            g.send_message(data.Source, data.Channel,
                "Informe as opções separando por \";\". Ex: Sim;Não voto;Talvez", data.Message.MessageId)
            return
        end
        g.print("[VOTE] Options recebido:", options)
        setar_opcoes(options)
        ev.data.ended = false
        ev.data.users_voted = {}
        ev.data.votacao_em_andamento = true
    end
end
