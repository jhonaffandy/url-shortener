package main

import (
	"log"
	"net/http"
	"url-shortener/handler"
	"url-shortener/repository"
)

// URL merepresentasikan satu record short URL.
type URL struct {
	Code        string `json:"code"`
	OriginalURL string `json:"original_url"`
}



// sebelum: initDB("urls.db")
// sesudah: initDB("data/urls.db")
func main() {
	db, err := repository.InitDB("data/urls.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	store := repository.NewURLStore(db)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.HealthHandler)
	mux.HandleFunc("POST /shorten", handler.ShortenHandler(store))
	mux.HandleFunc("GET /{code}", handler.RedirectHandler(store))
	mux.HandleFunc("GET /stats/{code}", handler.StatsHandler(store))

	log.Println("Server jalan di :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}