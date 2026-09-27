package main

import (
	"crypto/rand"
	"encoding/json"
	"log"
	"math/big"
	"net/http"
	"sync"
)

// URL merepresentasikan satu record short URL.
type URL struct {
	Code        string `json:"code"`
	OriginalURL string `json:"original_url"`
}

// store adalah "database" sementara kita, disimpan di memori.
// mutex dipakai supaya aman diakses dari banyak goroutine/request sekaligus.
type URLStore struct {
	mu   sync.Mutex
	data map[string]string // key: code, value: original URL
}

func NewURLStore() *URLStore {
	return &URLStore{
		data: make(map[string]string),
	}
}

func (s *URLStore) Save(code, url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[code] = url
}

func (s *URLStore) Get(code string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	url, ok := s.data[code]
	return url, ok
}

// generateShortCode bikin string random alfanumerik sepanjang `length`.
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateShortCode(length int) (string, error) {
	code := make([]byte, length)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		code[i] = charset[n.Int64()]
	}
	return string(code), nil
}

// --- HTTP Handlers ---

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
}

func shortenHandler(store *URLStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req shortenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Validasi kosong dulu di sini, validasi format URL beneran nyusul hari ke-4.
		if req.URL == "" {
			http.Error(w, "url is required", http.StatusBadRequest)
			return
		}

		code, err := generateShortCode(6)
		if err != nil {
			http.Error(w, "failed to generate code", http.StatusInternalServerError)
			return
		}

		store.Save(code, req.URL)

		resp := shortenResponse{
			Code:     code,
			ShortURL: "http://localhost:8080/" + code,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func redirectHandler(store *URLStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("code")

		originalURL, ok := store.Get(code)
		if !ok {
			http.Error(w, "short URL not found", http.StatusNotFound)
			return
		}

		http.Redirect(w, r, originalURL, http.StatusFound)
	}
}

func main() {
	store := NewURLStore()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("POST /shorten", shortenHandler(store))
	mux.HandleFunc("GET /{code}", redirectHandler(store))

	log.Println("Server jalan di :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}