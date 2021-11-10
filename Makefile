default: test

ci: test

test:
	go test -v ./... -coverprofile=coverage.out -covermode=count

integration:
	go test -v ./... -integration

lint:
	golangci-lint run ./...
