package mlua

import (
	"MyStreamBot/globals"
	"fmt"
	"reflect"

	lua "github.com/yuin/gopher-lua"
)

var (
	chatTable    *lua.LTable
	eventTable   *lua.LTable
	commandTable *lua.LTable
)

func ToLTable(L *lua.LState, data globals.MessageFromStream) *lua.LTable {
	if chatTable == nil {
		chatTable = L.NewTable()
	}
	chatTable.RawSetString("Source", lua.LString(data.Source))
	chatTable.RawSetString("Channel", lua.LString(data.Channel))
	chatTable.RawSetString("User", lua.LString(data.User))
	chatTable.RawSetString("UserId", lua.LString(data.UserId))
	chatTable.RawSetString("MessageId", lua.LString(data.MessageId))
	chatTable.RawSetString("Message", lua.LString(data.Message))
	metadata := L.NewTable()
	for k, v := range data.Metadata {
		metadata.RawSetString(k, lua.LString(fmt.Sprintf("%v", v)))
	}
	chatTable.RawSetString("Metadata", metadata)

	return chatTable
}

func ToLTableEvent(L *lua.LState, data globals.LuaEvent) *lua.LTable {
	if eventTable == nil {
		eventTable = L.NewTable()
	}
	eventTable.RawSetString("Type", lua.LString(data.Type))
	eventTable.RawSetString("User", lua.LString(data.User))
	eventTable.RawSetString("Text", lua.LString(data.Text))
	dataTable := L.NewTable()
	for k, v := range data.Data {
		dataTable.RawSetString(k, lua.LString(fmt.Sprintf("%v", v)))
	}
	eventTable.RawSetString("Data", dataTable)

	return eventTable
}

func ToLTableCommand(L *lua.LState, data globals.LuaCommand) *lua.LTable {
	if commandTable == nil {
		commandTable = L.NewTable()
	}
	commandTable.RawSetString("Name", lua.LString(data.Name))
	argsTable := L.NewTable()
	for _, arg := range data.Args {
		argsTable.Append(lua.LString(arg))
	}
	commandTable.RawSetString("Args", argsTable)
	commandTable.RawSetString("User", lua.LString(data.User))
	commandTable.RawSetString("Text", lua.LString(data.Text))
	dataTable := L.NewTable()
	for k, v := range data.Data {
		dataTable.RawSetString(k, lua.LString(fmt.Sprintf("%v", v)))
	}
	commandTable.RawSetString("Data", dataTable)
	commandTable.RawSetString("Source", lua.LString(data.Source))
	commandTable.RawSetString("Channel", lua.LString(data.Channel))
	commandTable.RawSetString("Message", ToLTable(L, data.Message))
	return commandTable
}

func StructToLTable(L *lua.LState, s interface{}) *lua.LTable {
	tbl := L.NewTable()
	v := reflect.ValueOf(s)
	t := reflect.TypeOf(s)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	if v.Kind() != reflect.Struct {
		return tbl
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Nome do campo
		key := fieldType.Name

		// Converter valor
		var lv lua.LValue
		switch field.Kind() {
		case reflect.String:
			lv = lua.LString(field.String())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			lv = lua.LNumber(field.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			lv = lua.LNumber(field.Uint())
		case reflect.Float32, reflect.Float64:
			lv = lua.LNumber(field.Float())
		case reflect.Bool:
			lv = lua.LBool(field.Bool())
		case reflect.Struct:
			lv = StructToLTable(L, field.Interface())
		default:
			lv = lua.LNil // tipos não suportados ainda
		}

		tbl.RawSetString(key, lv)
	}

	return tbl
}
