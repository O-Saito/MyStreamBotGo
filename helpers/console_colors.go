package helpers

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/metrics"
	"strings"
	"sync"
	"time"
)

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

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

type Metrics struct {
	Goroutines      int
	Alloc           uint64
	TotalAlloc      uint64
	Sys             uint64
	TotalCpuSeconds float64
}

var (
	infoLogger    *log.Logger
	errorLogger   *log.Logger
	debugLogger   *log.Logger
	warningLogger *log.Logger
	mu            sync.Mutex
	myMetrics     Metrics
)

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func (l Level) Color() string {
	switch l {
	case DEBUG:
		return Cyan
	case INFO:
		return Green
	case WARN:
		return Yellow
	case ERROR:
		return Red
	default:
		return Reset
	}
}

func Print(color string, message string) {
	log.Println(color + message + Reset)
}
func Printf(color string, format string, a ...any) {
	log.Printf(color+format+Reset+"\r\n", a...)
}

func Log(level Level, message string) {
	Print(level.Color(), message)
	SaveLog(level, message)
}
func Logf(level Level, format string, a ...any) {
	Printf(level.Color(), format, a...)
	SaveLog(level, fmt.Sprintf(format, a...))
}

func InitLog() error {
	err := os.MkdirAll("logs", 0755)
	if err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
	}
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFile, err := os.OpenFile(
		filepath.Join("logs", timestamp+".log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	mu.Lock()
	defer mu.Unlock()
	infoLogger = log.New(logFile, "INFO: ", log.Ldate|log.Ltime)
	errorLogger = log.New(logFile, "ERROR: ", log.Ldate|log.Ltime)
	debugLogger = log.New(logFile, "DEBUG: ", log.Ldate|log.Ltime)
	warningLogger = log.New(logFile, "WARNING: ", log.Ldate|log.Ltime)

	return nil
}

func SaveLog(level Level, message string) {
	mu.Lock()
	defer mu.Unlock()
	if debugLogger == nil || infoLogger == nil || warningLogger == nil || errorLogger == nil {
		log.Printf("Loggers not initialized: %s", message)
		return
	}
	const cpuMetric = "/cpu/classes/total:cpu-seconds"
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	samples := make([]metrics.Sample, 1)
	samples[0].Name = cpuMetric
	// Read the current cumulative CPU seconds
	metrics.Read(samples)
	if samples[0].Value.Kind() == metrics.KindBad {
		Logf(WARN, "Metric %s is not supported", cpuMetric)
	}

	goroutines := runtime.NumGoroutine()
	alloc := m.Alloc / 1024 / 1024
	totalAlloc := m.TotalAlloc / 1024 / 1024
	sys := m.Sys / 1024 / 1024
	totalCpuSeconds := samples[0].Value.Float64()

	if goroutines != myMetrics.Goroutines {
		myMetrics.Goroutines = goroutines
		infoLogger.Output(2, fmt.Sprintf("[METRICS] Goroutines: %d", goroutines))
	}
	if alloc != myMetrics.Alloc || totalAlloc != myMetrics.TotalAlloc || sys != myMetrics.Sys {
		myMetrics.Alloc = alloc
		myMetrics.TotalAlloc = totalAlloc
		myMetrics.Sys = sys
		infoLogger.Output(2, fmt.Sprintf("[METRICS] Mem Alloc: %d MiB, TotalAlloc: %d MiB, Sys: %d MiB", alloc, totalAlloc, sys))
	}
	if totalCpuSeconds != myMetrics.TotalCpuSeconds {
		myMetrics.TotalCpuSeconds = totalCpuSeconds
		infoLogger.Output(2, fmt.Sprintf("[METRICS] Total CPU Seconds: %f", totalCpuSeconds))
	}

	pc, file, line, _ := runtime.Caller(2)

	function := ""

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		function = "unknown"
	} else {
		function = fn.Name()
	}

	// The function name is typically in the format "packagepath.FunctionName"
	// Find the last dot to separate the package path from the function name
	lastDotIndex := strings.LastIndex(function, ".")
	if lastDotIndex != -1 {
		function = function[:lastDotIndex]
	}

	format := "[%s/%s:%d] %s\n"
	logLine := fmt.Sprintf(format, function, filepath.Base(file), line, message)

	switch level {
	case DEBUG:
		debugLogger.Output(2, logLine)
	case INFO:
		infoLogger.Output(2, logLine)
	case WARN:
		warningLogger.Output(2, logLine)
	case ERROR:
		errorLogger.Output(2, logLine)
	default:
	}
}
