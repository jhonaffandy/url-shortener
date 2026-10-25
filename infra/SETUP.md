# Setup VPS dari Kosong sampai Live

## 1. Akses awal
```bash
# Format: ssh root@<ip-vps>
ssh root@<IP>
```

## 2. Bikin user non-root
```bash
# Format: adduser <nama>		-> bikin user baru
# Format: usermod -aG <group> <nama>	-> tambah user ke group (sudo = boleh admin)
adduser deploy
usermod -aG sudo deploy
```

## 3. SSH key + hardening
```bash
# Di LAPTOP. Format: ssh-copy-id <user>@<ip>	-> pasang publi key laptop ke VPS
ssh-copy-id deploy@<ip>

# Di VPS. Edit config:
sudo nano /etc/ssh/sshd_config
#   PasswordAuthenticatio no
#   PermitRootLogin no
# Fromat: sudo sshd -t	-> cek syntax (nggak ada output = valid)
sudo sshd -t
sudo systemctl restart ssh

# Verifikasi (harus: no, yes, no):
sudo sshd -T | grep -Ei 'passwordauthentication|permitrootlogin|pubkeyauthentiaction'
```

PENTING: test login dari terminal BARU sebelum nutup terminal lama.

## 4. Firewall (UFW)
```bash
# Format: sudo ufw allow <port>		-> buka port
sudo ufw allow 22
sudo ufw allow 80
sudo ufw allow 443
sudo ufw enable
sudo ufw status verbose
```

Catatan: Docker bypass UFW untuk port yang di-publish. Cuma nginx yang boleh pakai `ports:` publik.
