#!/usr/bin/env bash
# 仅启动后端（调试用）
cd "$(dirname "$0")/backend"

if [ ! -d ".venv" ]; then
  python3 -m venv .venv
fi
source .venv/bin/activate

pip install -r requirements.txt -q

if [ ! -f ".env" ]; then
  cp .env.example .env
  echo "⚠️  已创建 .env 文件，请填写你的公众号 app_id 和 app_secret"
fi

echo "🚀 后端启动: http://127.0.0.1:8100"
uvicorn main:app --host 127.0.0.1 --port 8100 --reload
