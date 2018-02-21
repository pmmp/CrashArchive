all: run

build:
	go build -o ./bin/ca-pmmp

run: build
	./bin/ca-pmmp

build/linux:
	GOOS=linux go build -o ./bin/ca-pmmp-linux

cli/mysql:
	docker-compose exec db mysql -p -D crash_archive

defaultconfig:
	cp ./default-docker-compose.yml ./docker-compose.yml
	cp ./config/default-config.json ./config/config.json
