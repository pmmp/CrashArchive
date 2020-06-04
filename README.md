# CrashArchive

Web-based searchable archive for PocketMine-MP crash reports. https://crash.pmmp.io

## Setup in 30 seconds
CA is primarily used on Linux.

### Prerequisites
- Go 1.13+
- Docker

### Installing
Run the following:
```sh
git clone https://github.com/pmmp/CrashArchive
cd CrashArchive
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
