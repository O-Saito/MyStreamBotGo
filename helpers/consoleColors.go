package helpers

import "log"

var (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
	Kick   = "\033[38;2;255;69;0m"   // Kick orange color
	Twitch = "\033[38;2;100;65;165m" // Twitch purple color
	Lua    = "\033[38;2;0;128;255m"  // Lua blue color
)

func Log(color string, message string) {
	log.Println(color + message + Reset)
}
func Logf(color string, format string, a ...any) {
	log.Printf(color+format+Reset+"\r\n", a...)
}
