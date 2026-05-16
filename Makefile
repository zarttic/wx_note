.PHONY: dev build test clean

# ── 后端 ──────────────────────────────────────────────────────

dev-backend:
	cd backend && JWT_SECRET=dev-secret-do-not-use-in-production go run ./cmd/server/

build-backend:
	cd backend && go build -o wx-note ./cmd/server/

test-backend:
	cd backend && go test ./...

# ── 前端 ──────────────────────────────────────────────────────

dev-frontend:
	cd frontend && npm run dev

build-frontend:
	cd frontend && npm run build

# ── 组合 ──────────────────────────────────────────────────────

dev: dev-backend dev-frontend

build: build-backend build-frontend

test: test-backend
	cd frontend && npm test

clean:
	rm -f backend/wx-note
	rm -rf backend/data
