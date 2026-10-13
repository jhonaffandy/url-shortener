# Security Checklist (VPS + Project)

## A. Akses server (Minggu 6)
- [ ] Login pakai SSH key, bukan password
      Cek: `sudo sshd -T | grep passwordauthentication` -> harus `no`
- [ ] Nggak login/kerja sebagai root, pakai user biasa + sudo
      Cek: `whoami` -> bukan root
- [ ] Config SSH aman setelah edit
      Cek: `sudo sshd -t` -> nggak ada output = valid
      Catatan: tes login di terminal BARU sebelum tutup sesi lama

## B. Firewall (Minggu 7 H1-2)
- [ ] UFW aktif, default deny incoming, allow 22/80/443 saja
      Cek: `sudo ufw status verbose`
- [ ] Docker b ypass UFW. Cuma nginx yang boleh `ports:` (80/443)
      Service lain: tanpa `ports:` atau `127.0.0.1:port:port`
      Cek: `sudo ss -ltnp` -> cuma 22/80/443 di 0.0.0.0
- [ ] Firewall panel provider (MantaHost): allow 22, 80, 443 saja
      Cek: panel -> Servers -> Firewall -> harus ada 3 rule TCP + ICMP
      Catatan: ada 2 lapis firewall (panel + UFW), dua-duanya harus allow

## C. Secret & git (Minggu 7 H3)
- [ ] `.env` masuk `.gitignore`, `.env.example` di-commit
      Cek: `git check=ignore -v .env`
- [ ] Nggak ada secret di git history
      Cek: `git log --all --full-history -- .env` -> kosong
- [ ] Jangan hardcode password di compose.yml, pakai `.env`

## D. Update otomatis (Minggu 7 H4)
- [ ] unattended-upgrades aktif
      Cek: `cat /etc/apt/apt.conf.d/20auto-upgrades` -> dua-duanya "1"
- [ ] Origin security nggak dikomentari (`//`)
      Cek: `sudo unattended-upgrade --dry-run --debug 2>&1 | grep -i "allowed origins"` -> harus ada `a=noble-security`

## E. Pelajaran / jebakan
- `echo >>` bisa nyambung ke baris terakhir kalau file nggak berakhir newline -> pakai `printf "\n..."`
- Service "active" != config bener. Selalu cek isi config.
- `grep` case-sensitive, pakai `-i`. `-E` itu regex, bukan case.
- Sebelum matiin password login / ubah SSH: tes terminal baru dulu
