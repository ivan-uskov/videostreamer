ifneq (,$(wildcard ./.env))
    include .env
    include .env.local
    export
endif

install_apt_deps:
	sudo apt-get install -y v4l-utils
	sudo apt-get install -y libvpx-dev

dep:
	go mod tidy

lint:
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint golangci-lint run

fmt:
	go fmt ./src/...

build: fmt
	docker build . --force-rm

run:
	go run videostreamer/src/cmd