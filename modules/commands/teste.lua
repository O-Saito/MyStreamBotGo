function on_command(ev)
    g.print("[TESTE] Dados iniciais:", twitch.get_state())
    g.send_message(ev.Source, ev.Channel, "Comando de teste funcionando!", ev.Message.MessageId)
end