.PHONY: all clean deps build

GODEP := $(shell command -v dep 2> /dev/null)

ifndef GODEP
	go get -u github.com/golang/dep/cmd/dep
endif

all: clean deps build

deps:
	dep ensure -v

build:
	go build

clean:
	-rm -rf vendor barvin
