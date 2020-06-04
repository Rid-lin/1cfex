.PHONY :build

build:
	go build -o bin/1cfex.exe -v ./

.DUFAULT_GOAL := build

build_for_deploy:
	go build --ldflags "-w -s" -o bin/1cfex.exe -v ./

pack:
	upx.exe --ultra-brute bin/1cfex.exe

deploy: build_for_deploy pack