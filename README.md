# URL Shortener

Project untuk mengubah valid url yg panjang menjadi code 6 character. sehingga lebih mudah digunakan dibagikan dll

## Fitur

- Memendekkan URL jadi code 6 karakter (`POST /shorten`)
- Redirect ke URL asli lewat code (`GET /{code}`)
- Click counter tiap redirect
- Expiry link (opsional, kolom `expires_at`)
- Statistik link (`GET /stats/{code}`)

## Prasyarat

- Docker Desktop (include Docker Compose), atau
- Go 1.27+

## Cara Jalanin

### Pakai Docker (disarankan)

```bash
git clone <url-repo>
cd url-shortener
docker compose up -d --build
```

Server jalan di http://localhost:8080

### Tanpa Docker

```bash
git clone <url-repo>
cd url-shortener
mkdir -p data
go run .
```

## API

### `POST /shorten`

Request:

```bash
curl -X POST localhost:8080/shorten -d '{"url":"https://example.com"}'
```

Response 200:

```json
{ "code": "qWEwtj", "short_url": "http://localhost:8080/qWEwtj" }
```

### `GET /{code}`

Redirect 302 ke URL asli. 404 kalau code nggak ada atau sudah expired.

### `GET /stats/{code}`

Response 200:

```json
{
  "code": "qWEwtj",
  "clicks": 2,
  "created_at": "2026-10-10T13:54:05Z",
  "expires_at": null
}
```

### `GET /health`

Response 200: `{"status":"ok"}`

## Test

```bash
go test ./...
```

## Stop & Data

```bash
docker compose down      # stop, data tetap ada
docker compose down -v   # stop + HAPUS data (volume)
```

SQLite disimpan di named volume `urls_data` (`/app/data/urls.db`).
