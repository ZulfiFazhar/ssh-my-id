# SSH Portfolio - Setup & Maintenance Guide

## Server Setup

### Prerequisites

1. Linux server (Ubuntu/Debian)
2. Go 1.21+
3. Domain `ssh.zulfifazhar.dev` pointing to your server
4. SSH access to server

### Step 1: Install Dependencies

```bash
# Install nginx with stream module
sudo apt update
sudo apt install nginx-full

# Check nginx has stream module
nginx -V 2>&1 | grep stream
```

### Step 2: Configure SSH on Non-Standard Port

SSH daemon harus dipindah ke port lain karena port 22 dipakai nginx:

```bash
# Edit SSH config
sudo vim /etc/ssh/sshd_config
# Ubah: Port 22 → Port 2223

# Disable ssh.socket (Ubuntu modern)
sudo systemctl stop ssh.socket
sudo systemctl disable ssh.socket

# Restart SSH
sudo systemctl restart ssh
```

### Step 3: Configure Firewall

```bash
sudo ufw allow 2223/tcp  # SSH ke server
sudo ufw allow 22/tcp     # nginx proxy
```

### Step 4: Configure Nginx Stream Proxy

Edit `/etc/nginx/nginx.conf` - tambahkan block `stream` di luar `http` block:

```nginx
events {
    worker_connections 768;
}

http {
    # ... existing http config ...
}

stream {
    upstream ssh_backend {
        server 127.0.0.1:2222;
    }
    
    server {
        listen 22;
        proxy_pass ssh_backend;
        proxy_connect_timeout 10s;
    }
}
```

Restart nginx:
```bash
sudo nginx -t
sudo systemctl restart nginx
```

### Step 5: Deploy Application

```bash
# Clone repo
git clone https://github.com/yourusername/ssh-portfolio.git
cd ssh-portfolio

# Build
go build -o ssh-my-id

# Install binary
sudo cp ssh-my-id /home/ubuntu/app/ssh-my-id/

# Setup systemd service
sudo cp ssh-portfolio.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable ssh-portfolio
sudo systemctl start ssh-portfolio
```

## Update Code

### Pull Latest Changes

```bash
cd ~/app/ssh-my-id
git pull origin main

# Rebuild
go build -o ssh-my-id
```

### Restart Service

```bash
sudo systemctl restart ssh-portfolio
sudo systemctl status ssh-portfolio
```

## Troubleshooting

### SSH Connection Refused on Port 22

1. Check nginx is running:
   ```bash
   sudo systemctl status nginx
   ```

2. Check port binding:
   ```bash
   sudo ss -tlnp | grep :22
   ```

3. Verify app is running on port 2222:
   ```bash
   sudo ss -tlnp | grep 2222
   ```

### Cannot Connect via Domain

1. Check DNS is pointing to server:
   ```bash
   dig ssh.zulfifazhar.dev
   ```

2. Check firewall allows port 22:
   ```bash
   sudo ufw status
   ```

3. Test local connection:
   ```bash
   ssh -p 2223 localhost  # SSH ke server sendiri
   ```

### App Won't Start

1. Check logs:
   ```bash
   sudo journalctl -u ssh-portfolio -n 50
   ```

2. Check binary exists and is executable:
   ```bash
   ls -la /home/ubuntu/app/ssh-my-id/ssh-my-id
   file /home/ubuntu/app/ssh-my-id/ssh-my-id
   ```

3. Test binary manually:
   ```bash
   cd /home/ubuntu/app/ssh-my-id
   ./ssh-my-id
   ```

## File Structure

```
ssh-portfolio/
├── main.go           # Main application code
├── go.mod            # Go module definition
├── go.sum            # Go dependencies checksum
├── ssh-my-id         # Compiled binary
├── id_ed25519        # SSH key (private) - DO NOT COMMIT
├── id_ed25519.pub    # SSH key (public)
├── keys/             # Additional keys folder - DO NOT COMMIT
├── ssh-portfolio.service  # Systemd service file
├── docs/
│   └── README.md     # This file
└── README.md         # Project overview
```

## Security Notes

- Private SSH keys (`id_*`) are NEVER committed to git
- The `keys/` directory is gitignored
- App runs as root to bind privileged ports (22, 2222)
- SSH daemon runs on port 2223 to avoid port conflict
