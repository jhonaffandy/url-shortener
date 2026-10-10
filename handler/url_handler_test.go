package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"url-shortener/repository"
)

// TestHortenHandler_Success: POST url valid -> 200 + code 6 karakter + tersimpan di store
// Format: httptest.NewRequest(method, path string, body io.Reader)
func TestShortenHandler_Success(t *testing.T){
	db, err := repository.InitDB(":memory:") // ":memory:" = SQLite sementar di RAM, hilang pas test selesai
	if err != nil {
		t.Fatalf("initDB: %v", err)
	}
	store := repository.NewURLStore(db)
	handler := ShortenHandler(store) // balikin http.HandlerFunc

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
			db, err := repository.InitDB(":memory:") // ":memory:" = SQLite sementar di RAM, hilang pas test selesai
			if err != nil {
				t.Fatalf("initDB: %v", err)
			}
			handler := ShortenHandler(repository.NewURLStore(db))
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
// Skenario: stats code ada -> 200 + clicks sesuai
// Format: req.SetPathValue(name, value string)
// Contoh: req.SetPathValue("code", "abc123")
func TestStatsHandler_Found(t *testing.T) {
	db, err := repository.InitDB(":memory:")
	if err != nil {
		t.Fatalf("initDB: %v", err)
	}
	defer db.Close()
	store := repository.NewURLStore(db)

	if err := store.Save("abc123", "https://example.com"); err != nil {
		t.Fatalf("Save: %v", err)
	}
	store.IncrementClicks("abc123")

	req := httptest.NewRequest(http.MethodGet, "/stats/abc123", nil)
	req.SetPathValue("code", "abc123")
	rec := httptest.NewRecorder()

	StatsHandler(store)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got repository.StatsResponse
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
	db, err := repository.InitDB(":memory:")
	if err != nil {
		t.Fatalf("initDB: %v", err)
	}
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/stats/nggakada", nil)
	req.SetPathValue("code", "nggakada")
	rec := httptest.NewRecorder()

	StatsHandler(repository.NewURLStore(db))(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}