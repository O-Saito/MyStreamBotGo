function on_command(ev)
    g.send_message(ev.Source, ev.Channel, "https://livepix.gg/scavote", ev.Message.MessageId)
end