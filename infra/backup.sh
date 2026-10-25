#!/bin/bash
set -euo pipefail			# berhenti kalau ada command gagal / variabel kosong

VOLUME="url-shortener_urls_data"	# nama volume docker (hasil docker volume ls)
BACKUP_DIR="$HOME/backups"		# folder output
KEEP_DAYS=7				# backup lebih tua dari ini dihapus
FILE="urls-$(date +%Y%m%d-%H%M%S).db"	# contoh: urls-20261025-110900.db

mkdir -p "$BACKUP_DIR"

# Format: docker run --rm --user 0:0 -v <volume>:<path>:ro -v <folder-host:<path> <image> sqlite 3 <db> ".backup <output>"
docker run --rm --user 0:0 \
  -v "$VOLUME":/data:ro \
  -v "$BACKUP_DIR":/backup \
  keinos/sqlite3 \
  sqlite3 /data/urls.db ".backup /backup/$FILE"

# Format: sqlite3 <db> "PRAGMA integrity_check;" -> harus balikin "ok"
CHECK=$(sqlite3 "$BACKUP_DIR/$FILE" "PRAGMA integrity_check;")
if [ "$CHECK" != "ok" ]; then
  echo "integdridty check GAGAL: $CHECK" >&2
  exit 1
fi

# Format: gzip -c <input> > <output> (-c = tulis ke stdout, file asli nggak diubah)
gzip -c "$BACKUP_DIR/$FILE" > "$BACKUP_DIR/$FILE.gz"
rm -f "$BACKUP_DIR/$FILE"

# Format: finc <folder> -name '<pola>' -mtime +<hari> -delete
find "$BACKUP_DIR" -name 'urls-*.db.gz' -mtime +"$KEEP_DAYS" -delete

echo "OK: $BACKUP_DIR/$FILE.gz"
