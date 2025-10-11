function on_command(ev)
    g.log("debug teste", ev)
    --g.send_message(ev.Source, ev.Channel, "https://discord.gg/srWXN6KRsk")
    g.send_message(ev.Source, ev.Channel, "Comando de teste funcionando!", ev.Message.MessageId)
end