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

build_alpine: fmt
	docker build -f Dockerfile.alpine --force-rm -t ivanuskov/videostreamer:alpine .

build_debian: fmt
	docker build -f Dockerfile.debian --force-rm -t ivanuskov/videostreamer:debian .

build: build_alpine

run:
	go run videostreamer/src/cmd

run_docker:
	docker run -v /tmp/.X11-unix:/tmp/.X11-unix --ipc=host -e DISPLAY=$(DISPLAY) --privileged --env-file .env.local --device=/dev/video4 ivanuskov/videostreamer:alpine
