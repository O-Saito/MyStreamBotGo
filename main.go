package main

import (
	"MyStreamBot/globals"
	"MyStreamBot/goweb"
	"MyStreamBot/kick"
	"MyStreamBot/mlua"
	"MyStreamBot/sql"
	"MyStreamBot/twitch"
	"MyStreamBot/youtube"
	"runtime"
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
	}()*/

	go func() {
		// validar refresh tokens
	}()

	select {} // manter aplicação rodando
}
