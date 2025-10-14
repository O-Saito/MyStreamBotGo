package globals

import (
	"MyStreamBot/helpers"
	"bufio"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

type SocketMessage struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type MessageFromStream struct {
	Source    string         `json:"source"`
	Channel   string         `json:"channel"`
	UserId    string         `json:"userId"`
	User      string         `json:"user"`
	MessageId string         `json:"messageId"`
	Message   string         `json:"message"`
	Metadata  map[string]any `json:"metadata"`
}

type LuaEvent struct {
	Type string
	User string
	Text string
	Data map[string]interface{}
}

type LuaCommand struct {
	Source  string
	Channel string
	Name    string
	Args    []string
	User    string
	Text    string
	Message MessageFromStream
	Data    map[string]interface{}
}

type LuaChat struct {
	Source    string         `json:"source"`
	Channel   string         `json:"channel"`
	UserId    string         `json:"userId"`
	User      string         `json:"user"`
	MessageId string         `json:"messageId"`
	Message   string         `json:"message"`
	Metadata  map[string]any `json:"metadata"`
}

// WebSocket global exportado
var (
	WsBroadcast  = make(chan SocketMessage, 100)
	ChatQueue    = make(chan MessageFromStream, 200)
	CommandQueue = make(chan LuaCommand, 50)
	EventQueue   = make(chan LuaEvent, 100)
)
var sectionMap = map[string]any{
	"Config": GetConfig(),
	"State":  GetState(),
}

func LoadInitFile() {
	filePath := filepath.Join(".", "init.txt")
	file, err := os.Open(filePath)
	if err != nil {
		helpers.Logf(helpers.Red, "Erro ao abrir o arquivo: %v", err)
		os.WriteFile(filePath, []byte(""), 0644)
		return
	}
	defer file.Close() // Ensure the file is closed

	current := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		helpers.Logf(helpers.Cyan, "Lendo linha: %s", line)

		// Process the line (e.g., parse key-value pairs)
		if line == "" || line[0] == '#' {
			continue // Skip empty lines and comments
		}
		// Example processing (you can expand this as needed)
		if line[0] == '[' && line[len(line)-1] == ']' {
			current = line[1 : len(line)-1]
			helpers.Logf(helpers.Green, "Seção: %s", current)
			continue
		}
		if current == "" {
			helpers.Logf(helpers.Yellow, "Ignorando linha fora de seção: %s", line)
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			helpers.Logf(helpers.Yellow, "Linha inválida: %s", line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		obj, ok := sectionMap[current]
		if !ok {
			helpers.Logf(helpers.Yellow, "Seção desconhecida: %s", current)
			continue
		}

		// Caso especial: State.Data.algumacoisa=valor
		if current == "State" && strings.Contains(key, ".") {
			parts := strings.SplitN(key, ".", 2)
			if parts[0] == "Data" {
				GetState().Data[parts[1]] = value
				continue
			}
		}

		setField(obj, key, value)
	}

	if err := scanner.Err(); err != nil {
		helpers.Logf(helpers.Red, "Erro ao ler o arquivo: %v", err)
	}
}

// Define campos simples via reflection (ex: Config.BotPrefix)
func setField(obj any, field, value string) {
	v := reflect.ValueOf(obj).Elem()
	f := v.FieldByName(field)
	if !f.IsValid() || !f.CanSet() {
		helpers.Logf(helpers.Yellow, "Campo inválido: %s", field)
		return
	}

	switch f.Kind() {
	case reflect.String:
		f.SetString(value)
	case reflect.Bool:
		f.SetBool(strings.ToLower(value) == "true")
	case reflect.Int:
		// Conversão simples (sem erro fatal)
		if i, err := strconv.Atoi(value); err == nil {
			f.SetInt(int64(i))
		}
	default:
		helpers.Logf(helpers.Yellow, "Tipo de campo não suportado: %s", field)
	}
}
