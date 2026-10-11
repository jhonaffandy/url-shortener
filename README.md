# URL Shortener

Project untuk mengubah valid url yg panjang menjadi code 6 character. sehingga lebih mudah digunakan dibagikan dll

## Fitur

- Memendekkan URL jadi code 6 karakter (`POST /shorten`)
- Redirect ke URL asli lewat code (`GET /{code}`)
- Click counter tiap redirect
- Expiry link (opsional, kolom `expires_at`)
- Statistik link (`GET /stats/{code}`)

## Prasyarat

- Go 1.27+

## Cara Jalanin

git clone <url-repo>
cd url-shortener
go run .
​
Server jalan di `http://localhost:8080`. File `urls.db` (SQLite) dibuat otomatis.

## API

### POST /shorten

Request:
curl -X POST localhost:8080/shorten -d '{"url":"https://example.com"}'
​
Response 200:
​{"code":"qWEwtj","short_url":"http://localhost:8080/qWEwtj"}
​Error 400: URL kosong / bukan http(s) / JSON rusak.

### GET /{code}

Redirect 302 ke URL asli. 404 kalau code nggak ada atau sudah expired.

### GET /stats/{code}

Response 200:
{"code":"qWEwtj","clicks":2,"created_at":"2026-10-10T13:54:05Z","expires_at":null}
404 kalau code nggak ada.

### GET /health

Response 200: `{"status":"ok"}`

## Struktur Folder

- `main.go`: wiring route + jalanin server
- `handler/`: HTTP layer
- `service/`: validasi URL + generate code
- `repository/`: akses database (SQLite)

## Test

go test ./...
