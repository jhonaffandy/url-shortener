package repository

import "database/sql"

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
		"SELECT original_url FROM urls WHERE code = ? AND (expires_at IS NULL OR expires_at > datetime('now'))", 
		code,
	).Scan(&url)

	if err == sql.ErrNoRows{
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return url, true, nil
}



// Format: iniDB(path string) (*sql.DB, error)
// path = lokasi file sqlite
// return = koneksi db yang siap dipakai, atau error
// Contoh: db, err := initDB("urls.db")
func InitDB(path string) (*sql.DB, error){
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
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME,
		clicks INTEGER NOT NULL DEFAULT 0
	)`)

	if err != nil {
		return nil, err
	}

	return db, nil
}


// Format: type statsResponse struct { ... }
// Field : Code, Clicks, CreatedAt, ExpiresAt
// ExpiresAt pakai *string supaya NULL jadi null di JSON (bukan "")
type StatsResponse struct {
	Code		string 	`json:"code"`
	Clicks		int		`json:"clicks"`
	CreatedAt	string	`json:"created_at"`
	ExpiresAt	*string	`json:"expires_at"`
}

// Format: (s *URLStore) GetStats(code string) (*statsResponse, bool, error)
// Param : code = short code
// Return: (stats, ditemukan?, error)
// - ketemu       -> (&statsResponse{...}, true, nil)
// - nggak ketemu -> (nil, false, nil)
// - DB error     -> (nil, false, err)
// Contoh: st, ok, err := store.GetStats("abc123")
func (s *URLStore) GetStats(code string) (*StatsResponse, bool, error){
	var st StatsResponse
	var expires sql.NullString

	err := s.db.QueryRow(
		"SELECT code, clicks, created_at, expires_at FROM urls WHERE code = ?",
		code,
	).Scan(&st.Code, &st.Clicks, &st.CreatedAt, &expires)

	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	if expires.Valid{
		st.ExpiresAt = &expires.String
	}
	return &st, true, nil
}



// Format: (s *URLStore) IncrementClicks(code string) error
// - code = short code yang diklik
// - return error kalau query gagal
// Contoh: err := store.IncrementClicks("abc123")
func (s *URLStore) IncrementClicks(code string) error {
	_, err := s.db.Exec(
		"UPDATE urls SET clicks = clicks + 1 WHERE code = ?",
		code,
	)
	return err
}