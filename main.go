package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"url-shortener/repository"
	"url-shortener/service"
)

// URL merepresentasikan satu record short URL.
type URL struct {
	Code        string `json:"code"`
	OriginalURL string `json:"original_url"`
}



// --- HTTP Handlers ---

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
}



func shortenHandler(store *repository.URLStore) http.HandlerFunc {
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

		if !service.IsValidURL(req.URL){
			http.Error(w, "url must be a valid absolute URL (e.g. https://example.com)",http.StatusBadRequest)
			return
		}

		code, err := service.GenerateShortCode(6)
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

func redirectHandler(store *repository.URLStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("code")

		originalURL, ok, err := store.Get(code)
		if err != nil{
			log.Println("errrr")
			log.Printf("redirect Get error: %v", err)
			http.Error(w,"internal error", http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, "short URL not found", http.StatusNotFound)
			return
		}
		if err := store.IncrementClicks(code); err != nil {
			// Format: log.Printf(format string, args ...any)
			log.Printf("gagal increment clicks %s: %v", code, err)
		}

		http.Redirect(w, r, originalURL, http.StatusFound)
	}
}

func statsHandler(store *repository.URLStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request){
		code := r.PathValue("code")

		st, ok, err := store.GetStats(code)
		if err != nil {
			log.Printf("stats error: %v",err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return 
		}
		if !ok {
			http.Error(w, "short URL not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type","application/json")
		json.NewEncoder(w).Encode(st)
	}
}

func main() {
	db, err := repository.InitDB("urls.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	store := repository.NewURLStore(db)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("POST /shorten", shortenHandler(store))
	mux.HandleFunc("GET /{code}", redirectHandler(store))
	mux.HandleFunc("GET /stats/{code}", statsHandler(store))

	log.Println("Server jalan di :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}