package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func main() {
	// Format: sql.Open(driverName string, dataSourceName string) (*sql.DB, error)
	// driver name = "sqlite" (nama yang didaftarin driver)
	// dataSourceName = path file db
	// Contoh: db, err := sql.Open("sqlite", "test.db")
	db, err := sql.Open("sqlite", "test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Format: db.Ping() error
	// Cek koneksi beneran bisa. sql.Open sendiri belum konek
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("konek ke SQLite berhasil")
}