#!/bin/bash
# 为 magichub.top 申请 Let's Encrypt SSL 证书
# 运行前请确保 DNS 已解析到本服务器 (38.147.173.222)

set -e

echo "=== StaticMan SSL 证书申请 ==="
echo ""

# 检查 DNS
echo "检查 DNS 解析..."
MAGICHUB_IP=$(dig +short magichub.top 2>/dev/null || host magichub.top 2>/dev/null | awk '{print $NF}' | tail -1)
if [ "$MAGICHUB_IP" != "38.147.173.222" ]; then
    echo "⚠️  magichub.top 当前解析到: ${MAGICHUB_IP:-未解析}"
    echo "    需要解析到: 38.147.173.222"
    echo ""
    read -p "是否继续? [y/N] " confirm
    [[ "$confirm" != [yY] ]] && exit 1
fi

echo "✅ DNS 解析正确"
echo ""

# 申请证书
echo "申请 Let's Encrypt 证书..."
certbot certonly --standalone \
    -d magichub.top -d www.magichub.top \
    --agree-tos --no-eff-email \
    -m admin@magichub.top \
    --preferred-challenges http

echo ""
echo "✅ 证书申请成功"
echo ""

# 测试 Nginx
echo "测试并重载 Nginx..."
nginx -t && nginx -s reload

echo ""
echo "✅ 全部完成"
echo ""
echo "证书路径:"
echo "  /etc/letsencrypt/live/magichub.top/fullchain.pem"
echo "  /etc/letsencrypt/live/magichub.top/privkey.pem"
echo ""
echo "自动续期: certbot 已配置 systemd timer，无需手动操作"
