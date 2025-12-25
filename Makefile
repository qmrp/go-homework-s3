.PHONY: run test e2e

run:
	go run ./cmd/huayi-im

test:
	go test ./...

e2e: test
