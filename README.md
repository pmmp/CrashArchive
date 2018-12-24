# CrashArchive

Web-based searchable archive for PocketMine-MP crash reports. https://crash.pmmp.io

## Setup in 30 seconds
CA is primarily used on Linux.
Make sure you have Go 1.10 or newer installed in your PATH (and Docker), then run the following:
```sh
make defaultconfig
make build/linux
```
Tweak `docker-compose.yml` and `config.json` as you desire, and then run:
```sh
docker-compose up -d
```
