package console

import (
	"MyStreamBot/helpers"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type ConsoleEvent struct {
	Type string // "log", "chat", "metric"
	Data string
}

var ConsoleChan = make(chan ConsoleEvent, 1000)
var QuitChan = make(chan struct{})

const (
	maxLines     = 50 // Máximo de linhas no buffer
	metricHeight = 3  // Altura da área de métricas
)

// Public API
func Log(msg string, color ...string) {
	select {
	case ConsoleChan <- ConsoleEvent{Type: "log", Data: msg}:
	default:
		// descarta para não travar
	}
}

func Chat(user, msg string) {
	ConsoleChan <- ConsoleEvent{Type: "chat", Data: fmt.Sprintf("[%s]: %s", user, msg)}
}

func Metric(metric string) {
	ConsoleChan <- ConsoleEvent{Type: "metric", Data: metric}
}

// StartConsole inicializa a goroutine do console
func StartConsole() {
	go func() {
		helpers.Log(helpers.INFO, "Started console!")
		var lines []ConsoleEvent
		var metrics []string

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case ev := <-ConsoleChan:
				if ev.Type == "metric" {
					// Atualiza apenas a última métrica
					if len(metrics) < metricHeight {
						metrics = append(metrics, ev.Data)
					} else {
						metrics = append(metrics[1:], ev.Data)
					}
				} else {
					lines = append(lines, ev)
					if len(lines) > maxLines {
						lines = lines[len(lines)-maxLines:]
					}
				}
				drawConsole(lines, metrics)

			case <-ticker.C:
				// Atualiza métricas periodicamente
				cpu, mem := getCPUUsage(), getMemUsage()
				metricsText := fmt.Sprintf("CPU: %.1f%% | Mem: %.1f MB | Goroutines: %d", cpu, mem, runtime.NumGoroutine())
				if len(metrics) < metricHeight {
					metrics = append(metrics, metricsText)
				} else {
					metrics = append(metrics[1:], metricsText)
				}
				drawConsole(lines, metrics)

			case <-QuitChan:
				return
			}
		}
	}()
}

// Limpa o console
func clearScreen() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		fmt.Print("\033[H\033[2J")
	}
}

// Cores ANSI simples
func colorize(ev ConsoleEvent) string {
	switch ev.Type {
	case "log":
		return "\033[37m" + ev.Data + "\033[0m" // branco
	case "chat":
		return "\033[36m" + ev.Data + "\033[0m" // ciano
	case "metric":
		return "\033[33m" + ev.Data + "\033[0m" // amarelo
	default:
		return ev.Data
	}
}

// Desenha o console
func drawConsole(lines []ConsoleEvent, metrics []string) {
	clearScreen()

	fmt.Println("==== MyStreamBot Console ====")
	for _, m := range metrics {
		fmt.Println("\033[1m" + m + "\033[0m") // bold
	}
	fmt.Println(strings.Repeat("-", 50))

	for _, line := range lines {
		fmt.Println(colorize(line))
	}

	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("Scroll buffer, latest messages at bottom")
}

// StopConsole encerra a goroutine
func StopConsole() {
	close(QuitChan)
}

// --- Captura de CPU/Memória simples ---
func getCPUUsage() float64 {
	// Placeholder, pode ser aprimorado com pacotes OS
	return 0
}

func getMemUsage() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.Alloc) / 1024 / 1024
}
