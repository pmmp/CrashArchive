# CrashArchive

Web-based searchable archive for PocketMine-MP crash reports. https://crash.pmmp.io

## Setup in 30 seconds
CA is primarily used on Linux.

### Prerequisites
- Go 1.10+
- Docker

### Installing
Create a directory to install in, and run:
```sh
mkdir ca-pmmp && cd ca-pmmp
export GOPATH=$(pwd)
go get github.com/pmmp/CrashArchive
cd src/github.com/pmmp/CrashArchive
```
Run the following to generate configuration files:
```sh
make defaultconfig
```
Tweak `docker-compose.yml` and `config.json` as you desire, and then run:
```sh
make build/linux
docker-compose up -d
```
