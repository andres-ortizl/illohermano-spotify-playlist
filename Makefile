build:
	go build ./cmd/playlist/main.go

run: build
	./main