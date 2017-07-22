.PHONY: all clean deps build

GLIDE := $(shell command -v glide 2> /dev/null)

ifndef GLIDE
	go get -u github.com/Masterminds/glide
endif

all: clean deps build

deps:
	glide install

build:
	go build

clean:
	-rm -rf vendor barvin
