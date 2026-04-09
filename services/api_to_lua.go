package services

import (
	"MyStreamBot/helpers"
	"MyStreamBot/mlua"
	"reflect"
	"strings"
	"unicode"

	lua "github.com/yuin/gopher-lua"
)

type LuaFunction struct {
	Name string
	Fn   any
}

func patternizeName(name string) string {
	var result strings.Builder
	for i, r := range name {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func getFunctionName(lf LuaFunction) string {
	if lf.Name != "" {
		return lf.Name
	}
	return "unnamed"
}

func ExposeToLua(L *lua.LState, tableName string, functions []LuaFunction) {
	funcs := make(map[string]func(*lua.LState) int, len(functions))
	for _, lf := range functions {
		fnName := getFunctionName(lf)
		funcs[fnName] = buildLuaWrapper(lf.Fn)
	}
	mlua.ExposeServiceToLua(L, tableName, funcs)
	helpers.Printf(helpers.Green, "[API EXPOSED] %s with %d functions", tableName, len(functions))
}

func buildLuaWrapper(fn any) func(*lua.LState) int {
	fnType := reflect.TypeOf(fn)
	fnValue := reflect.ValueOf(fn)

	return func(L *lua.LState) int {
		numParams := fnType.NumIn()
		numArgs := L.GetTop()

		args := make([]reflect.Value, 0, numParams)
		argIndex := 0

		for i := range numParams {
			paramType := fnType.In(i)

			if paramType.Kind() == reflect.Slice && paramType.Elem().Kind() == reflect.String {
				var sliceArgs []string
				for argIndex < numArgs {
					if L.Get(argIndex+1) == lua.LNil {
						break
					}
					sliceArgs = append(sliceArgs, L.CheckString(argIndex+1))
					argIndex++
				}
				args = append(args, reflect.ValueOf(sliceArgs))
			} else if paramType.Kind() == reflect.Ptr && paramType.Elem().Kind() == reflect.String {
				if argIndex < numArgs && L.Get(argIndex+1) != lua.LNil {
					val := L.CheckString(argIndex + 1)
					args = append(args, reflect.ValueOf(&val))
					argIndex++
				} else {
					var nilPtr *string
					args = append(args, reflect.ValueOf(nilPtr))
				}
			} else {
				if argIndex < numArgs {
					converted := convertLuaToGo(L, argIndex, paramType)
					args = append(args, converted)
					argIndex++
				} else {
					args = append(args, reflect.Zero(paramType))
				}
			}
		}

		results := fnValue.Call(args)

		numResults := fnType.NumOut()
		if numResults == 0 {
			return 0
		}

		lastResult := results[numResults-1]
		if numResults >= 2 {
			if lastResult.Interface() != nil {
				if err, ok := lastResult.Interface().(error); ok {
					helpers.Logf(helpers.ERROR, "[LUA API] Error: %v", err)
					for i := 0; i < numResults-1; i++ {
						L.Push(lua.LNil)
					}
					L.Push(lua.LNil)
					return numResults - 1
				}
			}
		}

		for i := range numResults {
			result := results[i]
			if result.Kind() == reflect.Ptr {
				if result.IsNil() {
					L.Push(lua.LNil)
					continue
				}
				result = result.Elem()
			}
			lv := mlua.ToLValue(L, result.Interface())
			L.Push(lv)
		}

		return numResults
	}
}

func convertLuaToGo(L *lua.LState, argIndex int, targetType reflect.Type) reflect.Value {
	switch targetType.Kind() {
	case reflect.String:
		return reflect.ValueOf(L.CheckString(argIndex + 1))
	case reflect.Int:
		return reflect.ValueOf(int(L.CheckInt(argIndex + 1)))
	case reflect.Int8:
		return reflect.ValueOf(int8(L.CheckInt(argIndex + 1)))
	case reflect.Int16:
		return reflect.ValueOf(int16(L.CheckInt(argIndex + 1)))
	case reflect.Int32:
		return reflect.ValueOf(int32(L.CheckInt(argIndex + 1)))
	case reflect.Int64:
		return reflect.ValueOf(int64(L.CheckInt(argIndex + 1)))
	case reflect.Uint:
		return reflect.ValueOf(uint(L.CheckInt(argIndex + 1)))
	case reflect.Uint8:
		return reflect.ValueOf(uint8(L.CheckInt(argIndex + 1)))
	case reflect.Uint16:
		return reflect.ValueOf(uint16(L.CheckInt(argIndex + 1)))
	case reflect.Uint32:
		return reflect.ValueOf(uint32(L.CheckInt(argIndex + 1)))
	case reflect.Uint64:
		return reflect.ValueOf(uint64(L.CheckInt(argIndex + 1)))
	case reflect.Float32:
		return reflect.ValueOf(float32(L.CheckNumber(argIndex + 1)))
	case reflect.Float64:
		return reflect.ValueOf(float64(L.CheckNumber(argIndex + 1)))
	case reflect.Bool:
		return reflect.ValueOf(bool(L.CheckBool(argIndex + 1)))
	default:
		return reflect.Zero(targetType)
	}
}
