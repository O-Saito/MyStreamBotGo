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

func TableToMap(tbl *lua.LTable) map[string]interface{} {
	result := make(map[string]interface{})
	tbl.ForEach(func(key lua.LValue, value lua.LValue) {
		switch v := value.(type) {
		case lua.LString:
			result[key.String()] = string(v)
		case lua.LNumber:
			result[key.String()] = float64(v)
		case lua.LBool:
			result[key.String()] = bool(v)
		case *lua.LTable:
			result[key.String()] = TableToMap(v) // recursivo
		default:
			result[key.String()] = v.String()
		}
	})
	return result
}

func ToLValue(L *lua.LState, val any) lua.LValue {
	if val == nil {
		return lua.LNil
	}

	rv := reflect.ValueOf(val)
	rt := reflect.TypeOf(val)

	switch rv.Kind() {
	case reflect.String:
		return lua.LString(rv.String())
	case reflect.Bool:
		return lua.LBool(rv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return lua.LNumber(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return lua.LNumber(int64(rv.Uint()))
	case reflect.Float32, reflect.Float64:
		return lua.LNumber(rv.Float())
	case reflect.Map:
		tbl := L.NewTable()
		for _, key := range rv.MapKeys() {
			k := fmt.Sprint(key.Interface())
			v := rv.MapIndex(key).Interface()
			tbl.RawSetString(k, ToLValue(L, v))
		}
		return tbl
	case reflect.Slice, reflect.Array:
		tbl := L.NewTable()
		for i := 0; i < rv.Len(); i++ {
			elem := rv.Index(i).Interface()
			tbl.RawSetInt(i+1, ToLValue(L, elem))
		}
		return tbl
	case reflect.Struct:
		tbl := L.NewTable()
		for i := 0; i < rv.NumField(); i++ {
			field := rt.Field(i)
			// Ignora campos não exportados
			if field.PkgPath != "" {
				continue
			}
			name := field.Name
			/*if tag := field.Tag.Get("json"); tag != "" {
				tag = strings.Split(tag, ",")[0]
				if tag != "" && tag != "-" {
					name = tag
				}
			}*/
			fv := rv.Field(i).Interface()
			tbl.RawSetString(name, ToLValue(L, fv))
		}
		return tbl
	case reflect.Ptr:
		if !rv.IsNil() {
			return ToLValue(L, rv.Elem().Interface())
		}
		return lua.LNil
	default:
		helpers.Logf(helpers.ERROR, "[LUA PARSER] Default fallback %T", val)
		return lua.LString(fmt.Sprintf("%v", val))
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

func ToLTable(L *lua.LState, data *globals.MessageFromStream, tbl *lua.LTable) *lua.LTable {
	defer func() {
		if r := recover(); r != nil {
			helpers.Logf(helpers.ERROR, "Panic em ToLTable: %v", r)
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

func ToLTableEvent(L *lua.LState, data *globals.LuaEvent, tbl *lua.LTable) *lua.LTable {
	defer func() {
		if r := recover(); r != nil {
			helpers.Logf(helpers.ERROR, "Panic em ToLTableEvent: %v", r)
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

func ToLTableCommand(L *lua.LState, data *globals.LuaCommand, tbl *lua.LTable) *lua.LTable {
	defer func() {
		if r := recover(); r != nil {
			helpers.Logf(helpers.ERROR, "Panic em ToLTableCommand: %v", r)
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
	tbl.RawSetString("Message", ToLTable(L, &data.Message, tbl.RawGetString("Message").(*lua.LTable)))
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
