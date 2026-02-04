.PHONY: run test e2e

run:
	go run ./cmd/huayi-im/cmd/api/main.go

test:
	go test ./...

e2e: test
