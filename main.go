package main

import (

	//"MyStreamBot/eventsub"
	//"MyStreamBot/irc"
	//"MyStreamBot/lua"

	"MyStreamBot/globals"
	"MyStreamBot/goweb"
	"MyStreamBot/kick"
	"MyStreamBot/mlua"
	"MyStreamBot/twitch"
)

var (
	Version    = "dev"
	BuildDate  = "unknown"
	CommitHash = "none"
)

func main() {
	kick.Channels = []kick.IrcChannel{}
	twitch.Channels = []string{}

	globals.LoadInitFile()

	RegisterSocketHandlers()

	// Inicializa o package mlua
	mlua.Init(RegisterLuaFunctions)
	//mlua.ExposeFunctions()
	//mlua.RegisterGlobalState()

	// Carrega todos os módulos Lua e inicia hotreload
	mlua.LoadAllModules()
	mlua.StartWatcher()

	// Inicia goroutines de consumo das filas
	mlua.StartEventQueues()
	// iniciar servidor web
	goweb.StartHTTPServer()

	// iniciar login Twitch
	twitch.HandleLogin()
	kick.HandleLogin()

	// Para testes, simula mensagens de chat
	/*go func() {
		users := []string{"Alice", "Bob", "Carol"}
		for i := 0; ; i++ {
			ev := globals.MessageFromStream{
				Source:    "twitch",
				Channel:   "test_channel",
				User:      users[i%len(users)],
				UserId:    users[i%len(users)],
				MessageId: fmt.Sprintf("msgid-%d", i),
				Message:   fmt.Sprintf("Mensagem %d", i),
				Metadata:  nil,
			}
			select {
			case globals.ChatQueue <- ev:
			default:
				fmt.Println("[WARN] ChatQueue cheia, descartando mensagem")
			}
			time.Sleep(2 * time.Second)
		}
	}()*/

	select {} // manter aplicação rodando
}
