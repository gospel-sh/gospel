.PHONY: test bench

all: test

copyright:
	python3 .scripts/make_copyright_headers.py

test:
	go test ./...

bench:
	go test -bench=.
