#!/usr/bin/env bash
# wx_note 启动脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BACKEND_DIR="$SCRIPT_DIR/backend"
FRONTEND_DIR="$SCRIPT_DIR/frontend"

export PATH="/home/nefu_dev_lyj/go1.25/go/bin:$PATH"

# ─── Backend ───────────────────────────────────────────────────

echo "Building backend..."
cd "$BACKEND_DIR"
go build -o wx-note ./cmd/server/

echo "Starting backend (http://0.0.0.0:8100)..."
./wx-note &
BACKEND_PID=$!
sleep 2

# ─── Frontend ──────────────────────────────────────────────────

echo "Starting frontend (http://0.0.0.0:5173)..."
cd "$FRONTEND_DIR"
npm run dev &
FRONTEND_PID=$!
sleep 2

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  wx_note 已启动"
echo "  前端: http://0.0.0.0:5173"
echo "  后端: http://0.0.0.0:8100"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "按 Ctrl+C 停止"

trap "kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; exit" INT TERM
wait
