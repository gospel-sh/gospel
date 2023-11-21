.PHONY: test bench

all: test

test:
	go test ./...

bench:
	go test -bench=.
