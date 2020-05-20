.PHONY :build

build:
	go build -v ./

.DUFAULT_GOAL := build

pack:
	upx.exe --ultra-brute 1cfex.exe

deploy: build pack