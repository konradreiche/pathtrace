.PHONY: all build run clean

all: build

build:
	go build ./cmd/pathtrace

run: build
	./pathtrace
