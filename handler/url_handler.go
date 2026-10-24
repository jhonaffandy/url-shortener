package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"url-shortener/repository"
	"url-shortener/service"
)

// --- HTTP Handlers ---

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
}



func ShortenHandler(store *repository.URLStore) http.HandlerFunc {
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

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"version": "ci-test-1",
	})
}

func RedirectHandler(store *repository.URLStore) http.HandlerFunc {
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

func StatsHandler(store *repository.URLStore) http.HandlerFunc {
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
