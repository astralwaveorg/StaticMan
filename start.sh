#!/bin/bash
# MagicHub 本地启动脚本
set -e

cd "$(dirname "$0")"

export PORT="${PORT:-8080}"
export DATA_DIR="${DATA_DIR:-./data}"

# 检查必要文件
if [ ! -f "$DATA_DIR/password.yaml" ]; then
  echo "⚠️  $DATA_DIR/password.yaml 不存在，请创建密码配置"
fi

echo "🚀 启动 MagicHub..."
echo "   PORT:     $PORT"
echo "   DATA_DIR: $DATA_DIR"

# 启动 Go 服务（嵌入前端，使用 withweb 标签）
go run -tags withweb ./cmd/server/
