build:
	go build -o ./bin/go-grep .

run: build
	./bin/go-grep

test:
	go test ./...