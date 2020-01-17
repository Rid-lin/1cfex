.PHONY :build

build:
	go build -v ./

.DUFAULT_GOAL := build

pack:
	"D:\root\Go\bin\upx.exe" --ultra-brute 1cfex.exe