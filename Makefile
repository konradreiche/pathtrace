.PHONY: all build run clean

all: build

ui:
	cd ui/ && npm run build

build: ui
	go build ./cmd/pathtrace

run: build
	./pathtrace
