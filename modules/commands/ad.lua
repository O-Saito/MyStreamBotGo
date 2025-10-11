function on_command(ev)
    g.send_message(ev.Source, ev.Channel, "Hey não temos nenhum AD e infelizmente não sou parceiro da " .. ev.Source .. " mas cola ai no discord https://discord.gg/srWXN6KRsk", ev.Message.MessageId)
end