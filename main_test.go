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
	store := NewURLStore()
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
	original, ok := store.Get(resp.Code)
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
			handler := shortenHandler(NewURLStore())
			req := httptest.NewRequest(http.MethodPost,"/shorten", strings.NewReader((tc.body)))
			rec := httptest.NewRecorder()
			
			handler(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("body %s: status = %d, want %d", tc.body,rec.Code,http.StatusOK)
			}
		})
	}
}