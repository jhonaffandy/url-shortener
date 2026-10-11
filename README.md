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

```bash
git clone <url-repo>
cd url-shortener
go run .
```

## API

### POST /shorten

Request:

```bash
curl -X POST localhost:8080/shorten -d '{"url":"https://example.com"}'
```

Response 200:

```json
{ "code": "qWEwtj", "short_url": "http://localhost:8080/qWEwtj" }
```

### GET /stats/{code}

Response 200:

```json
{
  "code": "qWEwtj",
  "clicks": 2,
  "created_at": "2026-10-10T13:54:05Z",
  "expires_at": null
}
```

## Test

```bash
go test ./...
```
