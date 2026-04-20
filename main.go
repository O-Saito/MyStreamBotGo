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
	"os"
	"os/signal"
	"syscall"
)

var (
	Version    = "dev"
	BuildDate  = "unknown"
	CommitHash = "none"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			helpers.Logf(helpers.ERROR, "[MAIN] panic: %v", r)
		}
	}()

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

	//select {} // manter aplicação rodando
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 3. Gracefully shutdown with 10s timeout
	//ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()

	helpers.Logf(helpers.DEBUG, "Closing...")

	db.Close()
	mlua.SaveDynamicEvents()
	twitch.Close()
}
