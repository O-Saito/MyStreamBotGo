function on_start()
    print("[Lua] Evento de teste iniciado!")
    ev.setInterval(1)
    ev.setPaused(false)
    ev.data = { votos = { sim = 0, nao = 0 }, tempo = 10 }
end

function on_tick()
    local tempo = ev.data.tempo
    tempo = tempo - 1
    print("[Lua] Tempo restante:", tempo)
    ev.data.tempo = tempo

    if tempo <= 0 then
        print("[Lua] Votação encerrada!")
        ev.setPaused(true)
    end
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