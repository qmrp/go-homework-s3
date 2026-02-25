.PHONY: run run-backend run-frontend test e2e

run:
	@echo "Starting backend and frontend..."
	@trap 'kill 0' SIGINT; \
		(go run ./cmd/huayi-im/cmd/api/main.go) & \
		(pnpm -C im-app dev) & \
		wait

run-backend:
	go run ./cmd/huayi-im/cmd/api/main.go

run-frontend:
	pnpm -C im-app dev

test:
	go test ./...

e2e: test
