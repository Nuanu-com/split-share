

.PHONY: test
test:
	@go test ./... --v

.PHONY: vet
vet:
	@bad=$$(gofmt -l .); test -z "$$bad" || { echo "not gofmt'd:"; echo "$$bad" | sed 's/^/  /'; exit 1; }
	go vet ./...

.PHONY: cover
cover:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out
