# ToFix.md - Technical Debt and Refactoring Plan

## Project Overview: MyStreamBot
- **Type:** Streaming bot for Twitch/Kick/YouTube
- **Language:** Go 1.24.4
- **Total Go Files:** 34
- **Test Files:** 8
- **Total Tests:** 105 (all passing)

---

## Section 1: Critical Issues (Fix Now)

### 1.1 Panic Usage on Startup
**Locations:**
- `main.go:39` - DB connection failure - now returns error with `os.Exit(1)`
- `goweb/server.go:247` - InterfaceAddrs failure - now returns error with `os.Exit(1)`
- `goweb/server.go:261` - HTTP ListenAndServe - now properly handled

**Problem:** Process crashes without graceful error handling or retry logic.

**Status:** ✅ FIXED - All 3 panic locations now use `helpers.Logf()` + `os.Exit(1)` for graceful error handling

---

### 1.2 Ignored Error Handling (~36 instances)
**Locations:**
- `services/twitch/fetch.go` - 20 instances
- `services/youtube/fetch.go` - 4 instances
- `services/kick/fetch.go` - 5 instances
- `services/youtube/chat.go` - 1 instance
- `services/twitch/sub.go` - 2 instances
- `goweb/server.go` - 2 instances (non-critical, log and continue)
- `lua_functions.go` - 2 instances (non-critical, log and continue)

**Problem:** Silent failures with `_, _ := io.ReadAll()`, `_, _ := json.Unmarshal()`, `req, _ := http.NewRequest()` and `data, _ := json.Marshal()`

**Status:** ✅ FIXED - All instances now properly log and return errors
```go
// NOW PROPERLY HANDLED:
req, err := http.NewRequest("GET", url, nil)
if err != nil {
    helpers.Logf(helpers.ERROR, "[SERVICE] Function http.NewRequest failed: %v", err)
    return nil, err
}
body, err := io.ReadAll(resp.Body)
if err != nil {
    helpers.Logf(helpers.ERROR, "[SERVICE] Function io.ReadAll failed: %v", err)
    return nil, err
}
if err := json.Unmarshal(body, &u); err != nil {
    helpers.Logf(helpers.ERROR, "[SERVICE] Function json.Unmarshal failed: %v", err)
    return nil, err
}
```

### 1.3 Portuguese Logs to English
**Problem:** ~80+ Portuguese log messages throughout codebase

**Status:** ✅ FIXED - All Portuguese messages translated to English
- "Erro ao..." → "Error when..."
- "Falha ao..." → "Failed to..."
- "Token não encontrado" → "Token not found"
- "Canal não encontrado" → "Channel not found"
- "Conexão..." → "Connection..."
- "Seção:" → "Section:"
- "Campo inválido" → "Invalid field"
- "Login concluído" → "Login completed"
- etc.

---

## Section 2: High Priority Issues

### 2.1 Missing Test Coverage by Package

| Package | Coverage | Status | Priority |
|---------|----------|--------|----------|
| `helpers/` | 20.6% | Low | HIGH |
| `twitch/` | 14.7% | Low | HIGH |
| `kick/` | 23.1% | Low | HIGH |
| `youtube/` | 31.5% | Medium | MEDIUM |
| `globals/` | 45.6% | Medium | MEDIUM |
| `sql/` | 67.8% | High | LOW |
| `processors/` | 66.0% | High | LOW |
| **`goweb/`** | **0%** | **No Tests** | **HIGH** |
| **`mlua/`** | **0%** | **No Tests** | **HIGH** |

**Recommended Fix:**
1. Create tests for `goweb/` WebSocket server - 8 tests needed
2. Create tests for `mlua/` Lua integration - 10 tests needed

---

### 2.2 Code Quality Issues

#### Unused Code
- `console/console.go` - **DELETED** (confirmed unused, removed dead code)
- `helpers/console_colors.go` - All code is actively used (308 references)

#### Unhandled Type Assertions (multiple)
**Locations:** handlers.go, goweb/server.go, services/twitch/twitch.go, services/twitch/sub.go, services/kick/kick.go, services/youtube/chat.go

**Problem:** `data["key"].(string)` without proper type checking

**Status:** ✅ FIXED - All type assertions converted to comma-ok idiom with proper error logging

---

## Section 3: Medium Priority Issues

### 3.1 Architecture Concerns

#### Global Mutable State
- `globals.ChatQueue`, `globals.CommandQueue`, etc. - Global channels without protection
- Makes testing harder and creates hidden dependencies

**Current Pattern:**
```go
var ChatQueue = make(chan MessageFromStream, 200)
```

**Consider:** Dependency injection for better testability

---

### 3.2 Missing Error Context

Many API functions return generic errors:
```go
return fmt.Errorf("user not found")
```

**Status:** ✅ FIXED - All 21 error messages now include function name and parameters for context

---

### 3.3 Logging Inconsistency
- Some places use `helpers.Log()`, others use `fmt.Println()`
- No structured logging

**Status:** ✅ FIXED - Changed `fmt.Printf` in helpers/console_colors.go to `helpers.Logf()`

---

## Section 4: Security Concerns

### 4.1 Insecure Config Storage
- `init.txt` contains credentials in plain text
- No encryption for stored tokens

### 4.2 WebSocket Security
- `goweb/server.go:19` - `CheckOrigin` returns always true:
```go
CheckOrigin: func(r *http.Request) bool {
    return true
}
```

**Should be:** Verify origin against allowed domains

---

## Section 5: Recommended Test Structure

### 5.1 Tests Needed by Package

#### helpers/ (5 additional tests)
- Test `Contains` edge cases (nil, empty)
- Test `Find` with no match
- Test `GenerateCodeChallenge` edge cases

#### globals/ (3 additional tests)
- Test `LoadInitFile` with invalid sections
- Test `LoadTwitchSubTypes` with malformed JSON

#### goweb/ (new package - 8 tests)
- `TestSocketConnection`
- `TestSocketMessageParsing`
- `TestBroadcastFiltering`
- `TestUpgradeConnection`
- `TestAdminDelete`
- `TestAdminBan`
- `TestAdminAutomod`

#### mlua/ (new package - 10 tests)
- `TestToLValue_string`
- `TestToLValue_struct`
- `TestToLValue_map`
- `TestFromLValue_table`
- `TestTableToMap`

---

## Section 6: Implementation Order

### Phase 1: Critical Fixes (1 week)
- [x] Replace panic calls with proper error handling
- [x] Fix all ignored error handling (~36 instances)
- [x] Translate Portuguese logs to English (~87 messages)

### Phase 2: Test Coverage (2 weeks)
- [ ] Add tests for goweb/ package
- [ ] Add tests for mlua/ package

### Phase 3: Code Quality (1 week)
- [x] Fix type assertions without checks
- [x] Add error context to API functions
- [x] Remove unused console code OR add tests

### Phase 4: Security (1 week)
- [ ] Fix WebSocket origin check
- [ ] Document credential storage requirements

---

## Summary Statistics

| Metric | Value |
|--------|-------|
| Total Go Files | 34 |
| Test Files | 8 |
| Tests (Passing) | 105 |
| Overall Coverage | ~30% |
| **Critical Issues** | **0** |
| High Priority Issues | 10+ |
| Medium Priority Issues | 15+ |
| Security Issues | 2 (Config Storage, WebSocket Origin) |

---

## Validation Results (Last Updated: 2026-04-22)

### Section 1 - Critical Issues: ALL RESOLVED ✅

| Issue | Status | Evidence |
|------|--------|----------|
| 1.1 Panic Usage | ✅ FIXED | 0 panic() calls in codebase |
| 1.2 Ignored Errors | ✅ FIXED | 0 `_ := func()` in production code |
| 1.3 Portuguese Logs | ✅ FIXED | 0 Portuguese strings found |

**All three critical issues have been fully resolved.** Build and tests pass successfully.