deps:
	rm -f go.*
	go mod init github.com/353solutions/journal
	go mod tidy

test:
	golangci-lint run ./...
	gosec ./...
	go test -v