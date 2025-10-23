package mlua

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"fmt"
	"reflect"

	lua "github.com/yuin/gopher-lua"
)

var (
// chatTable    *lua.LTable
// eventTable   *lua.LTable
// commandTable *lua.LTable
)

func ToLValue(L *lua.LState, val any) lua.LValue {
	switch v := val.(type) {
	case nil:
		return lua.LNil
	case string:
		return lua.LString(v)
	case bool:
		return lua.LBool(v)
	case int:
		return lua.LNumber(v)
	case int64:
		return lua.LNumber(v)
	case float32:
		return lua.LNumber(v)
	case float64:
		return lua.LNumber(v)
	case map[string]any:
		tbl := L.NewTable()
		for k, vv := range v {
			tbl.RawSetString(k, ToLValue(L, vv))
		}
		return tbl
	case []any:
		tbl := L.NewTable()
		for i, vv := range v {
			tbl.RawSetInt(i+1, ToLValue(L, vv)) // Lua é 1-index
		}
		return tbl
	default:
		return lua.LString(fmt.Sprintf("%v", v)) // fallback
	}
}

func FromLValue(L *lua.LState, lv lua.LValue) any {
	switch v := lv.(type) {
	/*case lua.LNilType:
	return nil*/
	case lua.LBool:
		return bool(v)
	case lua.LNumber:
		return float64(v) // ou int se você quiser forçar
	case lua.LString:
		return string(v)
	case *lua.LTable:
		// Decide se é map ou slice
		// Checa se existem índices numéricos sequenciais
		max := 0
		isArray := true
		v.ForEach(func(key, value lua.LValue) {
			if k, ok := key.(lua.LNumber); ok {
				if int(k) > max {
					max = int(k)
				}
			} else {
				isArray = false
			}
		})

		if isArray && max > 0 {
			arr := make([]any, max)
			i := 0
			v.ForEach(func(_, value lua.LValue) {
				arr[i] = FromLValue(L, value)
				i++
			})
			return arr
		}

		// Caso contrário, é um mapa
		m := make(map[string]any)
		v.ForEach(func(key, value lua.LValue) {
			m[fmt.Sprint(key)] = FromLValue(L, value)
		})
		return m
	default:
		return fmt.Sprintf("%v", v) // fallback
	}
}

func ToLTable(L *lua.LState, data globals.MessageFromStream, tbl *lua.LTable) *lua.LTable {
	defer func() {
		if r := recover(); r != nil {
			helpers.Logf(helpers.Red, "Panic em ToLTable: %v", r)
		}
	}()
	tbl.RawSetString("Source", lua.LString(data.Source))
	tbl.RawSetString("Channel", lua.LString(data.Channel))
	tbl.RawSetString("User", lua.LString(data.User))
	tbl.RawSetString("UserId", lua.LString(data.UserId))
	tbl.RawSetString("MessageId", lua.LString(data.MessageId))
	tbl.RawSetString("Message", lua.LString(data.Message))
	metadata := L.NewTable()
	for k, v := range data.Metadata {
		metadata.RawSetString(k, lua.LString(fmt.Sprintf("%v", v)))
	}
	tbl.RawSetString("Metadata", metadata)

	return tbl
}

func ToLTableEvent(L *lua.LState, data globals.LuaEvent, tbl *lua.LTable) *lua.LTable {
	defer func() {
		if r := recover(); r != nil {
			helpers.Logf(helpers.Red, "Panic em ToLTableEvent: %v", r)
		}
	}()
	tbl.RawSetString("Type", lua.LString(data.Type))
	//tbl.RawSetString("User", lua.LString(data.User))
	//tbl.RawSetString("Text", lua.LString(data.Text))
	dataTable := L.NewTable()
	for k, v := range data.Data {
		dataTable.RawSetString(k, lua.LString(fmt.Sprintf("%v", v)))
	}
	tbl.RawSetString("Data", dataTable)

	return tbl
}

func ToLTableCommand(L *lua.LState, data globals.LuaCommand, tbl *lua.LTable) *lua.LTable {
	defer func() {
		if r := recover(); r != nil {
			helpers.Logf(helpers.Red, "Panic em ToLTableCommand: %v", r)
		}
	}()

	tbl.RawSetString("Name", lua.LString(data.Name))
	argsTable := L.NewTable()
	for _, arg := range data.Args {
		argsTable.Append(lua.LString(arg))
	}
	tbl.RawSetString("Args", argsTable)
	tbl.RawSetString("User", lua.LString(data.User))
	tbl.RawSetString("Text", lua.LString(data.Text))
	dataTable := L.NewTable()
	for k, v := range data.Data {
		dataTable.RawSetString(k, lua.LString(fmt.Sprintf("%v", v)))
	}
	tbl.RawSetString("Data", dataTable)
	tbl.RawSetString("Source", lua.LString(data.Source))
	tbl.RawSetString("Channel", lua.LString(data.Channel))
	if _, ok := tbl.RawGetString("Message").(*lua.LTable); !ok {
		tbl.RawSetString("Message", L.NewTable())
	}
	tbl.RawSetString("Message", ToLTable(L, data.Message, tbl.RawGetString("Message").(*lua.LTable)))
	return tbl
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
		case reflect.Slice, reflect.Array:
			arr := L.NewTable()
			rv := reflect.ValueOf(v)
			for i := 0; i < rv.Len(); i++ {
				arr.Append(StructToLTable(L, rv.Index(i).Interface()))
			}
			return arr
		default:
			lv = lua.LNil // tipos não suportados ainda
		}

		tbl.RawSetString(key, lv)
	}

	return tbl
}
