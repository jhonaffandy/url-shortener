package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strings"

	_ "modernc.org/sqlite"
)

// URL merepresentasikan satu record short URL.
type URL struct {
	Code        string `json:"code"`
	OriginalURL string `json:"original_url"`
}

// store adalah "database" sementara kita, disimpan di memori.
// mutex dipakai supaya aman diakses dari banyak goroutine/request sekaligus.
type URLStore struct {
	db *sql.DB
}

// Format: NewURLStore(db *sql.DB) *URLStore
func NewURLStore(db *sql.DB) *URLStore {
	return &URLStore{
		db: db,
	}
}

// Format: (s *URLStore) Save(code, url string) error
// Simpan 1 baris ke table urls. Return error kalau gagal (misal code duplikat)
// Contoh: err := store.Save("abc123", "https://example.com")
func (s *URLStore) Save(code, url string) error {
	_, err := s.db.Exec(
		"INSERT INTO URLS (code, original_url) VALUES (?,?)",
		code, url,
	)
	return err
}

// Format: (s *URLStore) Get(code string) (string, bool, error)
// Return: (url, ditemukan?, error)
// - ketemu -> ("https://..", true, nill)
// - nggak ketemu -> (("", false, nil)
// - DB error -> ("", false, err)
// Contoh: url, ok, err := store.Get("abc123") 
func (s *URLStore) Get(code string) (string, bool, error) {
	var url string

	// Format: db.QueryRow(query string, args ...any) *aql.Row
	// Format: row.Scan(dest ...any) error  -> isi variabel dari kolom hasil (pakai &)
	err := s.db.QueryRow(
		"SELECT original_url FROM urls WHERE code = ?", code,
	).Scan(&url)

	if err == sql.ErrNoRows{
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return url, true, nil
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


func isValidURL(raw string) bool {
	u, err:= url.ParseRequestURI(raw)
	if err != nil{
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https"{
		return false
	}
	if u.Host == ""{
		return false
	}
	return true
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

		req.URL = strings.TrimSpace(req.URL)

		if req.URL == "" {
			http.Error(w, "url is required", http.StatusBadRequest)
			return
		}

		if !isValidURL(req.URL){
			http.Error(w, "url must be a valid absolute URL (e.g. https://example.com)",http.StatusBadRequest)
			return
		}

		code, err := generateShortCode(6)
		if err != nil {
			http.Error(w, "failed to generate code", http.StatusInternalServerError)
			return
		}

		if err := store.Save(code, req.URL); err != nil {
			http.Error(w,"failed to save url", http.StatusInternalServerError)
			return
		}

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

		originalURL, ok, err := store.Get(code)
		if err != nil{
			http.Error(w,"internal error", http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, "short URL not found", http.StatusNotFound)
			return
		}

		http.Redirect(w, r, originalURL, http.StatusFound)
	}
}

// Format: iniDB(path string) (*sql.DB, error)
// path = lokasi file sqlite
// return = koneksi db yang siap dipakai, atau error
// Contoh: db, err := initDB("urls.db")
func initDB(path string) (*sql.DB, error){
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err:= db.Ping(); err != nil {
		return nil, err
	}

	// Format: db.Exec(query string, args ..anuy) (sql.Result, error)
	// Dipakai untuk statement yang nggak balikin b aris (CREATE, INSERT, UPDATE, DELETE)
	// Contoh: _, err := db.Exec("DELETE FROM urls WHERE code = ?", "abc123")
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS URLS (
			code TEXT PRIMARY KEY,
			original_url TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)

	if err != nil {
		return nil, err
	}

	return db, nil
}

func main() {
	db, err := initDB("urls.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	store := NewURLStore(db)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("POST /shorten", shortenHandler(store))
	mux.HandleFunc("GET /{code}", redirectHandler(store))

	log.Println("Server jalan di :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}