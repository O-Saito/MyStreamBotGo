package main

import (
	"MyStreamBot/globals"
	"MyStreamBot/goweb"
	"MyStreamBot/helpers"
	"MyStreamBot/kick"
	"MyStreamBot/mlua"
	"MyStreamBot/sql"
	"MyStreamBot/twitch"
	"MyStreamBot/youtube"
)

var (
	Version    = "dev"
	BuildDate  = "unknown"
	CommitHash = "none"
)

func main() {
	kick.Channels = []kick.IrcChannel{}
	twitch.Channels = []string{}

	helpers.InitLog()

	globals.LoadInitFile()
	globals.LoadTwitchSubTypes()

	db, err := sql.NewCoreDB(globals.GetConfig().DBName)
	if err != nil {
		panic(err)
	}
	globals.SetGlobalDB(db)

	RegisterSocketHandlers()

	// Inicializa o package mlua
	mlua.Init(RegisterLuaFunctions, RegisterTwitchLuaFunctions)

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
	youtube.HandleLogin()

	/*go func() {
		for {
			mem := getMemUsage()
			helpers.Logf(helpers.Blue, "Mem: %.1f MB | Goroutines: %d", mem, runtime.NumGoroutine())
			time.Sleep(5 * time.Second)
		}
	}()

	go func() {
		// validar refresh tokens
	}()*/

	select {} // manter aplicação rodando
}
