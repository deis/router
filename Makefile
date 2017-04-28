include includes.mk

SHORT_NAME := router
DEIS_REGISTRY ?= ${DEV_REGISTRY}
IMAGE_PREFIX ?= deis

include versioning.mk

SHELL_SCRIPTS = $(wildcard rootfs/bin/*) rootfs/opt/router/sbin/boot

REPO_PATH := github.com/deis/${SHORT_NAME}

# The following variables describe the containerized development environment
# and other build options
DEV_ENV_IMAGE := quay.io/deis/go-dev:v0.22.0
DEV_ENV_WORK_DIR := /go/src/${REPO_PATH}
DEV_ENV_CMD := docker run --rm -v ${CURDIR}:${DEV_ENV_WORK_DIR} -w ${DEV_ENV_WORK_DIR} ${DEV_ENV_IMAGE}
DEV_ENV_CMD_INT := docker run -it --rm -v ${CURDIR}:${DEV_ENV_WORK_DIR} -w ${DEV_ENV_WORK_DIR} ${DEV_ENV_IMAGE}
LDFLAGS := "-s -w -X main.version=${VERSION}"
BINDIR := ./rootfs/opt/router/sbin

# The following variables describe the source we build from
GO_FILES := $(wildcard *.go)
GO_DIRS := model/ nginx/ utils/ utils/modeler
GO_PACKAGES := ${REPO_PATH} $(addprefix ${REPO_PATH}/,${GO_DIRS})

# The binary compression command used
UPX := upx -9 --mono --no-progress

# The following variables describe k8s manifests we may wish to deploy
# to a running k8s cluster in the course of development.
DEPLOYMENT := manifests/deis-${SHORT_NAME}-deployment.yaml
SVC := manifests/deis-${SHORT_NAME}-service.yaml

# Allow developers to step into the containerized development environment
dev: check-docker
	${DEV_ENV_CMD_INT} bash

# Containerized dependency resolution
bootstrap: check-docker
	${DEV_ENV_CMD} glide install

# Containerized build of the binary
build: check-docker
	mkdir -p ${BINDIR}
	${DEV_ENV_CMD} make binary-build

docker-build: build check-docker
	docker build ${DOCKER_BUILD_FLAGS} -t ${IMAGE} rootfs
	docker tag ${IMAGE} ${MUTABLE_IMAGE}

# Builds the binary-- this should only be executed within the
# containerized development environment.
binary-build:
	GOOS=linux GOARCH=amd64 go build -o ${BINDIR}/${SHORT_NAME} -ldflags ${LDFLAGS} ${SHORT_NAME}.go
	$(call check-static-binary,$(BINDIR)/${SHORT_NAME})
	${UPX} ${BINDIR}/${SHORT_NAME}

deploy: check-kubectl docker-build docker-push
	kubectl --namespace=deis patch deployment deis-${SHORT_NAME} \
		--type='json' \
		-p='[ \
			{"op": "replace", "path": "/spec/strategy", "value":{"type":"Recreate"}}, \
			{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value":"$(IMAGE)"}, \
			{"op": "replace", "path": "/spec/template/spec/containers/0/imagePullPolicy", "value":"Always"} \
		]'

test: test-style test-unit test-functional

test-cover:
	${DEV_ENV_CMD} test-cover.sh

test-functional:
	@echo no functional tests

test-style: check-docker
	${DEV_ENV_CMD} make style-check

# This should only be executed within the containerized development environment.
style-check:
	lint
	shellcheck $(SHELL_SCRIPTS)

test-unit:
	${DEV_ENV_CMD} go test --cover --race -v ${GO_PACKAGES}
