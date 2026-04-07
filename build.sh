#!/usr/bin/env bash
set -e

echo "Building pi-sentinel for Linux ARM64..."
export PATH="/c/Program Files/Go/bin:$PATH"
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o pi-sentinel .

echo "Copying to Pi..."
scp -i ~/.ssh/pi_key pi-sentinel gueva@192.168.1.45:/usr/local/bin/pi-sentinel

echo ""
echo "Done. On the Pi, run:"
echo "  sudo nano /etc/systemd/system/pi-sentinel.service"
echo ""
echo "Paste this content:"
cat <<'EOF'
[Unit]
Description=Pi Sentinel - system monitor
After=network.target

[Service]
ExecStart=/usr/local/bin/pi-sentinel daemon
Restart=always
RestartSec=10
User=gueva
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
echo ""
echo "Then run:"
echo "  sudo systemctl daemon-reload"
echo "  sudo systemctl enable pi-sentinel"
echo "  sudo systemctl start pi-sentinel"
echo "  sudo systemctl status pi-sentinel"
echo ""
echo "Set your Telegram token:"
echo "  pi-sentinel set telegram.bot_token YOUR_TOKEN"
echo "Add services to watch:"
echo "  pi-sentinel set services.watchlist nginx,myapp"
