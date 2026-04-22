package sql

import (
	"testing"
	"time"
)

func TestCoreDB_SaveToken_GetToken(t *testing.T) {
	db, err := NewCoreDB("test_token")
	if err != nil {
		t.Fatalf("NewCoreDB() error = %v", err)
	}
	defer db.db.Close()

	expiresAt := time.Now().Add(1 * time.Hour)

	err = db.SaveToken("twitch", "access_token_value", "refresh_token_value", expiresAt)
	if err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}

	token, err := db.GetToken("twitch")
	if err != nil {
		t.Fatalf("GetToken() error = %v", err)
	}

	if token.AccessToken != "access_token_value" {
		t.Errorf("GetToken().AccessToken = %q, want %q", token.AccessToken, "access_token_value")
	}
	if token.RefreshToken != "refresh_token_value" {
		t.Errorf("GetToken().RefreshToken = %q, want %q", token.RefreshToken, "refresh_token_value")
	}
}

func TestCoreDB_GetToken_NotFound(t *testing.T) {
	db, err := NewCoreDB("test_notfound")
	if err != nil {
		t.Fatalf("NewCoreDB() error = %v", err)
	}
	defer db.db.Close()

	_, err = db.GetToken("nonexistent")
	if err == nil {
		t.Error("GetToken() for nonexistent should return error")
	}
}

func TestCoreDB_DeleteToken(t *testing.T) {
	db, err := NewCoreDB("test_delete")
	if err != nil {
		t.Fatalf("NewCoreDB() error = %v", err)
	}
	defer db.db.Close()

	expiresAt := time.Now().Add(1 * time.Hour)
	db.SaveToken("twitch", "access", "refresh", expiresAt)

	err = db.DeleteToken("twitch")
	if err != nil {
		t.Fatalf("DeleteToken() error = %v", err)
	}

	_, err = db.GetToken("twitch")
	if err == nil {
		t.Error("GetToken() after delete should return error")
	}
}

func TestCoreDB_KVSet_KVGet(t *testing.T) {
	db, err := NewCoreDB("test_kv")
	if err != nil {
		t.Fatalf("NewCoreDB() error = %v", err)
	}
	defer db.db.Close()

	err = db.KVSet("key1", "value1")
	if err != nil {
		t.Fatalf("KVSet() error = %v", err)
	}

	val, err := db.KVGet("key1")
	if err != nil {
		t.Fatalf("KVGet() error = %v", err)
	}

	if val != "value1" {
		t.Errorf("KVGet(%q) = %q, want %q", "key1", val, "value1")
	}
}

func TestCoreDB_KVSet_Overwrite(t *testing.T) {
	db, err := NewCoreDB("test_overwrite")
	if err != nil {
		t.Fatalf("NewCoreDB() error = %v", err)
	}
	defer db.db.Close()

	db.KVSet("key1", "value1")
	db.KVSet("key1", "value2")

	val, _ := db.KVGet("key1")
	if val != "value2" {
		t.Errorf("KVGet(%q) after overwrite = %q, want %q", "key1", val, "value2")
	}
}

func TestCoreDB_KVGet_NotFound(t *testing.T) {
	db, err := NewCoreDB("test_notfoundkv")
	if err != nil {
		t.Fatalf("NewCoreDB() error = %v", err)
	}
	defer db.db.Close()

	_, err = db.KVGet("nonexistent")
	if err == nil {
		t.Error("KVGet() for nonexistent should return error")
	}
}

func TestCoreDB_KVDelete(t *testing.T) {
	db, err := NewCoreDB("test_kvdelet")
	if err != nil {
		t.Fatalf("NewCoreDB() error = %v", err)
	}
	defer db.db.Close()

	db.KVSet("key1", "value1")
	err = db.KVDelete("key1")
	if err != nil {
		t.Fatalf("KVDelete() error = %v", err)
	}

	_, err = db.KVGet("key1")
	if err == nil {
		t.Error("KVGet() after delete should return error")
	}
}

func TestCoreDB_KVSet_ComplexValue(t *testing.T) {
	db, err := NewCoreDB("test_complex")
	if err != nil {
		t.Fatalf("NewCoreDB() error = %v", err)
	}
	defer db.db.Close()

	complexVal := map[string]any{
		"name":   "test",
		"count": 42,
		"active": true,
	}

	err = db.KVSet("complex", complexVal)
	if err != nil {
		t.Fatalf("KVSet() error = %v", err)
	}

	val, err := db.KVGet("complex")
	if err != nil {
		t.Fatalf("KVGet() error = %v", err)
	}

	m, ok := val.(map[string]any)
	if !ok {
		t.Fatalf("KVGet() returned wrong type: %T", val)
	}
	if m["name"] != "test" || m["count"].(float64) != 42 {
		t.Errorf("KVGet() complex value = %+v", m)
	}
}

func TestCoreDB_SaveToken_Update(t *testing.T) {
	db, err := NewCoreDB("test_update")
	if err != nil {
		t.Fatalf("NewCoreDB() error = %v", err)
	}
	defer db.db.Close()

	expires1 := time.Now().Add(1 * time.Hour)
	db.SaveToken("twitch", "access1", "refresh1", expires1)

	expires2 := time.Now().Add(2 * time.Hour)
	db.SaveToken("twitch", "access2", "refresh2", expires2)

	token, _ := db.GetToken("twitch")
	if token.AccessToken != "access2" {
		t.Errorf("SaveToken() update failed: got %q", token.AccessToken)
	}
}