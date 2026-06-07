#!/bin/bash
# StaticMan 一键构建+部署脚本
# 用法: ./scripts/deploy.sh [服务器地址]
# 默认服务器: root@38.147.173.222

set -euo pipefail

SERVER="${1:-root@38.147.173.222}"
PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BINARY="$PROJECT_DIR/staticman-linux"
APP_DIR="/opt/magichub"
BIN_NAME="staticman"
SERVICE="staticman.service"
HEALTH_TIMEOUT=30

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info()  { echo -e "${BLUE}[INFO]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_ok()    { echo -e "${GREEN}[OK]${NC} $1"; }
log_err()   { echo -e "${RED}[ERR]${NC} $1"; }

# ==================== 本地构建 ====================
cd "$PROJECT_DIR"

log_info "步骤 1/5: 构建前端..."
cd "$PROJECT_DIR/web"
npm run build

cd "$PROJECT_DIR"
log_info "步骤 2/5: 编译后端 (Linux AMD64, embed模式)..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags withweb -o staticman-linux ./cmd/server

# 验证
if ! file "$BINARY" | grep -q "ELF.*x86-64"; then
    log_err "编译失败: 不是 Linux x86-64 ELF"
    exit 1
fi

if ! strings "$BINARY" | grep -q "dist/assets/index-.*\.js"; then
    log_warn "⚠️ 警告: 二进制可能未嵌入前端资源"
    log_warn "   确认使用了 -tags withweb 编译"
    exit 1
fi

log_ok "编译成功: $(ls -lh "$BINARY" | awk '{print $5}')"

# ==================== 上传部署 ====================
log_info "步骤 3/5: 部署到服务器 ($SERVER)..."

# 备份旧版本
ssh "$SERVER" "
    if [ -f $APP_DIR/bin/$BIN_NAME ]; then
        cp $APP_DIR/bin/$BIN_NAME $APP_DIR/bin/${BIN_NAME}.bak.\$(date +%s)
        ls -t $APP_DIR/bin/${BIN_NAME}.bak.* 2>/dev/null | tail -n +6 | xargs -r rm -f
    fi
"

# 停止服务
log_info "停止服务..."
ssh "$SERVER" "systemctl stop $SERVICE || true"
sleep 1

# 上传
scp "$BINARY" "$SERVER:$APP_DIR/bin/${BIN_NAME}.new"
ssh "$SERVER" "chmod +x $APP_DIR/bin/${BIN_NAME}.new && mv $APP_DIR/bin/${BIN_NAME}.new $APP_DIR/bin/$BIN_NAME"

# 清理旧开发模式目录
ssh "$SERVER" "rm -rf $APP_DIR/src/internal/web/dist"

# 启动服务
log_info "启动服务..."
ssh "$SERVER" "systemctl start $SERVICE"

# ==================== 健康检查 ====================
log_info "步骤 4/5: 健康检查..."
sleep 2

for i in $(seq 1 $HEALTH_TIMEOUT); do
    STATUS=$(ssh "$SERVER" "curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:8080/ 2>/dev/null || echo '000'")
    if [ "$STATUS" = "200" ] || [ "$STATUS" = "401" ] || [ "$STATUS" = "302" ]; then
        log_ok "服务健康 (HTTP $STATUS)"
        break
    fi
    if [ "$i" -eq "$HEALTH_TIMEOUT" ]; then
        log_err "健康检查失败! HTTP $STATUS"
        log_info "查看日志: ssh $SERVER 'journalctl -u $SERVICE -n 50'"
        exit 1
    fi
    sleep 1
done

# 验证前端版本
REMOTE_JS=$(ssh "$SERVER" "curl -s http://127.0.0.1:8080/ | grep -o 'index-[A-Za-z0-9]*\.js' | head -1")
LOCAL_JS=$(strings "$BINARY" | grep -o 'dist/assets/index-[A-Za-z0-9]*\.js' | head -1 | sed 's|dist/assets/||')
if [ "$REMOTE_JS" = "$LOCAL_JS" ]; then
    log_ok "前端版本一致: $REMOTE_JS"
else
    log_warn "前端版本可能不一致 (远程: $REMOTE_JS, 本地: $LOCAL_JS)"
fi

# ==================== 完成 ====================
log_info "步骤 5/5: 完成"
echo ""
log_ok "✅ 部署完成!"
echo ""
echo "访问地址:"
echo "  https://files.magichub.top"
echo "  https://magichub.top (DNS配置后生效)"
echo ""
echo "服务状态:"
ssh "$SERVER" "systemctl status $SERVICE --no-pager -l | head -6"
