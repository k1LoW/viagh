default: test

ci: test

test:
	go test -v ./...

integration:
	go test -v ./... -integration

lint:
	golangci-lint run ./...
