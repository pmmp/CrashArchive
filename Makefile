all: run

build:
	go build -o ./bin/ca-pmmp

run: build
	./bin/ca-pmmp

build/linux:
	GOOS=linux go build -o ./bin/ca-pmmp-linux
