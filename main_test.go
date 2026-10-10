package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIsValidURL(t *testing.T){
	got := isValidURL("https://example.com")
	if got != true {
		t.Errorf("isValidURL(%q) = %v, mau %v","https://example.com",got,true)
	}

	got = isValidURL("bukan-url")
	if got != false {
		t.Errorf("isValidURL(%q) = %v, mau %v", "bukan-url",got,false)
	}
}

func TestGenerateShortCode(t *testing.T){
	length := 6

	got, err := generateShortCode(length)
	// cek main error fail all fatal
	if err!=nil {
		t.Fatalf("generateShortCode(%d) error: %v",length,err)
	}
	// cek panjang
	if len(got) != length {
		t.Errorf("len(generateShortCode(%d)) = %d, want %d", length,len(got),length)
	}
	// cel tiap karakter ad di charset
	for _, c:= range got {
		if !strings.ContainsRune(charset,c){
			t.Errorf("karakter %q di kode %q tidak ad di charset", c, got)
		}
	}
}


func TestGenerateShortCodeUnique(t *testing.T){
	a, err := generateShortCode(6)
	if err != nil{
		t.Fatalf("error: %v", err)
	}
	b, err := generateShortCode(6)
	if err!= nil {
		t.Fatalf("error: %v", err)
	}

	if a==b{
		t.Errorf("dua kode sama: %q, harusnya beda", a)
	}

}

// TestHortenHandler_Success: POST url valid -> 200 + code 6 karakter + tersimpan di store
// Format: httptest.NewRequest(method, path string, body io.Reader)
func TestShortenHandler_Success(t *testing.T){
	db, err := initDB(":memory:") // ":memory:" = SQLite sementar di RAM, hilang pas test selesai
	if err != nil {
		t.Fatalf("initDB: %v", err)
	}
	store := NewURLStore(db)
	handler := shortenHandler(store) // balikin http.HandlerFunc

	body := strings.NewReader(`{"url":"https://example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/shorten", body)
	rec := httptest.NewRecorder()

	handler(rec, req) // panggil langsung: handler(w,r)

	// cek status
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code,http.StatusOK)
	}

	// decode response JSON ke struct shrtenResponse
	var resp shortenResponse
	if err:=json.NewDecoder(rec.Body).Decode(&resp);err != nil{
		t.Fatalf("decode response: %v", err)
	}

	if len(resp.Code) != 6 {
		t.Errorf("len(code) = %d, want 6", len(resp.Code))
	}

	// cek tersimpan di store: Get(code) (string, bool)
	original, ok , err := store.Get(resp.Code)
	if !ok{
		t.Fatalf("code %q tidak ad di store", resp.Code)
	}
	if original != "https://example.com"{
		t.Errorf("store.Get(%q) = %q, want %q", resp.Code, original, "https://example.com")
	}
}

// TestShortenhandler_BadRequest: input salah -> 400
func TestShortenHandler_BadRequest(t *testing.T){
	// struct anonim: 1 baris = 1 skenario
	tests := []struct{
		name string // nama skenario
		body string // isi request
	}{
		{"url kosong", `{"url":""}`},
		{"url cuma spasi", `{"url":"   "}`},
		{"bukan url", `{"url":"bukan-url"}`},
		{"scheme ftp", `{"url":"ftp://example.com"}`},
		{"json rusak", `{url:`},
	}

	for _,tc := range tests {
		// t.Run(name string, func(t *testing.T)) -> subtest, muncul sendiri di output
		t.Run(tc.name, func(t *testing.T){
			db, err := initDB(":memory:") // ":memory:" = SQLite sementar di RAM, hilang pas test selesai
			if err != nil {
				t.Fatalf("initDB: %v", err)
			}
			handler := shortenHandler(NewURLStore(db))
			req := httptest.NewRequest(http.MethodPost,"/shorten", strings.NewReader((tc.body)))
			rec := httptest.NewRecorder()
			
			handler(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("body %s: status = %d, want %d", tc.body,rec.Code,http.StatusBadRequest)
			}
		})
	}
}

// Format: TestXxx(t *testing.T)
// Skenario: link expried -> Get return ok=false
func TestGet_Expired(t *testing.T){
	db, err := initDB(":memory:")
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
	db, err := initDB(":memory:")
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

// Format: TestXxx(t *testing.T)
// Skenario: stats code ada -> 200 + clicks sesuai
// Format: req.SetPathValue(name, value string)
// Contoh: req.SetPathValue("code", "abc123")
func TestStatsHandler_Found(t *testing.T) {
	db, err := initDB(":memory:")
	if err != nil {
		t.Fatalf("initDB: %v", err)
	}
	defer db.Close()
	store := NewURLStore(db)

	if err := store.Save("abc123", "https://example.com"); err != nil {
		t.Fatalf("Save: %v", err)
	}
	store.IncrementClicks("abc123")

	req := httptest.NewRequest(http.MethodGet, "/stats/abc123", nil)
	req.SetPathValue("code", "abc123")
	rec := httptest.NewRecorder()

	statsHandler(store)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got statsResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Clicks != 1 {
		t.Errorf("clicks = %d, want %d", got.Clicks, 1)
	}
	if got.ExpiresAt != nil {
		t.Errorf("expires_at = %v, want nil", got.ExpiresAt)
	}
}

// Skenario: stats code nggak ada -> 404
func TestStatsHandler_NotFound(t *testing.T) {
	db, err := initDB(":memory:")
	if err != nil {
		t.Fatalf("initDB: %v", err)
	}
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/stats/nggakada", nil)
	req.SetPathValue("code", "nggakada")
	rec := httptest.NewRecorder()

	statsHandler(NewURLStore(db))(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}