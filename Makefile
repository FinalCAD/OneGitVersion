GOFILES = $(shell find . -name '*.go')
GOMAIN = ./cmd/git_version_one

default: build

workdir:
	mkdir -p workdir

build: workdir/one-git-version

build-native: $(GOFILES)
	go build -o workdir/one-git-version $(GOMAIN)

workdir/one-git-version: $(GOFILES)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o workdir/one-git-version $(GOMAIN)