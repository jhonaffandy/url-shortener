package repository

import (
	"testing"
)

// Format: TestXxx(t *testing.T)
// Skenario: link expried -> Get return ok=false
func TestGet_Expired(t *testing.T){
	db, err := InitDB(":memory:")
	if err != nil {
		t.Fatalf("iniDB: %v", err)
	}
	defer db.Close()
	store := NewURLStore(db)

	// Format: db.Exec(query string, args ...any) (sql.Result, error)
	_, err = db.Exec(
		"INSERT INTO urls (code, original_url, expires_at) VALUES (?,?, datetime('now','-1 hour'))",
		"lewat1", "https://a.com",
	)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	_, ok, err := store.Get("lewat1")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if ok {
		t.Errorf("Get(%q) ok = %v, want %v", "lewat1", ok, false)
	}
}

// Format: TestXxx(t *testing.T)
// Skenario: link belum expired -> Get return ok=true + URL asli
func TestGet_NotExpired(t *testing.T) {
	db, err := InitDB(":memory:")
	if err != nil {
		t.Fatalf("initDB: %v", err)
	}
	defer db.Close()
	store := NewURLStore(db)

	_, err = db.Exec(
		"INSERT INTO urls (code, original_url, expires_at) VALUES (?, ?, datetime('now', '+1 hour'))",
		"masih1", "https://b.com",
	)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	got, ok, err := store.Get("masih1")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if !ok {
		t.Fatalf("Get(%q) ok = %v, want %v", "masih1", ok, true)
	}
	if got != "https://b.com" {
		t.Errorf("Get(%q) = %q, want %q", "masih1", got, "https://b.com")
	}
}
