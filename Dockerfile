# -------- modules --------
FROM golang:1.18 AS modules

COPY ./go.mod ./go.sum /
RUN go mod download

# -------- build binary --------
FROM golang:1.18 AS builder

RUN apt-get update
RUN apt-get install -y v4l-utils
RUN apt-get install -y libvpx-dev

RUN useradd -u 1001 appuser

COPY --from=modules /go/pkg /go/pkg
COPY . /build
WORKDIR /build

RUN GOARCH=amd64 GOOS=linux CGO_ENABLED=0 \
    go build -o ./bin/videostreamer ./src/cmd

RUN chmod +x ./bin/videostreamer

# -------- build image --------
FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /etc/passwd /etc/passwd
USER appuser

COPY --from=builder /build/bin/videostreamer /app/bin/videostreamer
COPY ./.env /app/bin/.env

WORKDIR /app

CMD ["/app/bin/videostreamer"]