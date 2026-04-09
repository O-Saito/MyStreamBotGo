package main

import (
	"MyStreamBot/globals"
	"MyStreamBot/goweb"
	"MyStreamBot/helpers"
	"MyStreamBot/mlua"
	"MyStreamBot/processors"
	"MyStreamBot/services/kick"
	"MyStreamBot/services/twitch"
	"MyStreamBot/services/youtube"
	"MyStreamBot/sql"
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
	mlua.Init(RegisterLuaFunctions, RegisterServiceAPIs)

	// Carrega todos os módulos Lua e inicia hotreload
	mlua.LoadAllModules()
	mlua.StartWatcher()

	// Inicia goroutines de consumo das filas
	go processors.ProcessChatQueue()
	go processors.ProcessCommandQueue()
	go processors.ProcessEventQueue()
	go processors.ProcessDyEventQueue()
	go processors.ProcessLuaRequest()

	// iniciar servidor web
	goweb.StartHTTPServer()

	// iniciar login Twitch
	twitch.HandleLogin()
	kick.HandleLogin()
	youtube.HandleLogin()

	select {} // manter aplicação rodando
}
