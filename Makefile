.PHONE: devs, devc
devs:
	CONF_FILE_PATH=./config/dev.yaml go run main.go server

devc:
	CONF_FILE_PATH=./config/dev.yaml go run main.go client

APP_NAME := isayme/go-xlan
APP_VERSION := $(shell git describe --tags --always)
APP_PKG := $(shell echo ${PWD} | sed -e "s\#${GOPATH}/src/\#\#g")
BUILD_TIME := $(shell date -u +"%FT%TZ")
GIT_REVISION := $(shell git rev-parse HEAD)

.PHONY: build
build:
	go build -ldflags "-X ${APP_PKG}/xlan/util.Version=${APP_VERSION} \
	-X ${APP_PKG}/xlan/util.BuildTime=${BUILD_TIME} \
	-X ${APP_PKG}/xlan/util.GitRevision=${GIT_REVISION}" \
	-o ./dist/xlan main.go

.PHONY: image
image:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make build
	docker build --rm -t ${APP_NAME}:${APP_VERSION} .

.PHONY: publish
publish: image
	docker push ${APP_NAME}:${APP_VERSION}
	docker tag ${APP_NAME}:${APP_VERSION} ${APP_NAME}:latest
	docker push ${APP_NAME}:latest
