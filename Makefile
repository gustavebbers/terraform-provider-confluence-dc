default: build

.PHONY: build
build:
	go install .

.PHONY: test
test:
	go test -v -cover -timeout=120s ./...

.PHONY: testacc
testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: generate
generate:
	go generate ./...

.PHONY: fmt
fmt:
	gofmt -s -w .
