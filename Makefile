BINARY_NAME = totalmix-mac-volume-key

.PHONY: build

build:
	go build -o $(BINARY_NAME) .
