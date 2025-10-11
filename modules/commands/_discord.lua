function on_command(ev)
    g.log("debug-discord", ev)
    g.send_message(ev.Source, ev.Channel, "https://discord.gg/srWXN6KRsk", ev.Message.MessageId)
    --g.send_message(ev.Source, ev.Channel, "https://discord.gg/srWXN6KRsk")
end