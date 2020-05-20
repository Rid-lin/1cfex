.PHONY :build

build:
	go build -o bin/1cfex.exe -v ./

.DUFAULT_GOAL := build

pack:
	upx.exe --ultra-brute 1cfex.exe

deploy: build pack