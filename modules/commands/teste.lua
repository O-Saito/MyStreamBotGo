function on_command(ev)
    g.send_message(ev.Source, ev.Channel, "Comando de teste funcionando!", ev.Message.MessageId)
end