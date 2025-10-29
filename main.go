package main

import (

	//"MyStreamBot/eventsub"
	//"MyStreamBot/irc"
	//"MyStreamBot/lua"

	"MyStreamBot/globals"
	"MyStreamBot/goweb"
	"MyStreamBot/helpers"
	"MyStreamBot/kick"
	"MyStreamBot/mlua"
	"MyStreamBot/twitch"
	"runtime"
	"time"
)

var (
	Version    = "dev"
	BuildDate  = "unknown"
	CommitHash = "none"
)

func getMemUsage() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.Alloc) / 1024 / 1024
}

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

	go func() {
		mem := getMemUsage()
		helpers.Logf(helpers.Blue, "Mem: %.1f MB | Goroutines: %d", mem, runtime.NumGoroutine())
		time.Sleep(5 * time.Second)
	}()

	select {} // manter aplicação rodando
}
