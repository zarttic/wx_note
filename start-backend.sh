#!/usr/bin/env bash
# 仅启动后端（调试用）
cd "$(dirname "$0")/backend"

echo "🔨 构建后端..."
go build -o wx-note ./cmd/server/

echo "🚀 后端启动: http://127.0.0.1:${PORT:-8100}"
echo "   请确保已设置 JWT_SECRET 环境变量"
echo ""
./wx-note
