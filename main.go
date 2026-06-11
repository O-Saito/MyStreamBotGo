package main

import (
	"MyStreamBot/globals"
	"MyStreamBot/goweb"
	"MyStreamBot/helpers"
	"MyStreamBot/mlua"
	"MyStreamBot/plugin"
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
		helpers.Logf(helpers.ERROR, "[MAIN] Failed to connect to database: %v", err)
		helpers.Log(helpers.ERROR, "[MAIN] Exiting due to unrecoverable error")
		os.Exit(1)
	}
	globals.SetGlobalDB(db)

	RegisterSocketHandlers()

	plugin.InitAll()

	// Initialize mlua package
	mlua.Init(RegisterLuaFunctions, RegisterServiceAPIs, plugin.RegisterLuaActions)

	// Load all Lua modules and start hotreload
	mlua.LoadAllModules()
	mlua.StartWatcher()

	// Start queue consumer goroutines
	go processors.ProcessChatQueue()
	go processors.ProcessCommandQueue()
	go processors.ProcessEventQueue()
	go processors.ProcessDyEventQueue()
	go processors.ProcessLuaRequest()

	plugin.StartAll()

	// start web server
	goweb.StartHTTPServer()

	go twitch.RestoreSession()
	go youtube.RestoreSession()
	go kick.RestoreSession()

	// start Twitch login
	twitch.HandleLogin()
	youtube.HandleLogin()
	kick.HandleLogin()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	helpers.Logf(helpers.DEBUG, "Closing...")

	plugin.StopAll()
	db.Close()
	mlua.SaveDynamicEvents()
	twitch.Close()
}
