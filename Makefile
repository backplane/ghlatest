.PHONY: build lint clean test

build: lint test ghlatest

ghlatest: *.go */*.go
	@echo '==> Building $@'
	go build -o "$@"

lint:
	@echo '==> Linting'
	go fmt
	go vet
	staticcheck

test:
	@echo '==> Testing'
	go test -v

clean:
	@echo '==> Cleaning'
	rm -rf -- ghlatest test

ftest: clean ghlatest
	@echo '==> Doing some functional checks'
	mkdir test
	(cd test && ../ghlatest dl --current-os --filter all --extract --keep snakeeyes --rm glvnst/snakeeyes && ls -al)
	(cd test && ../ghlatest dl --current-arch --filter linux --extract --keep ghlatest --rm backplane/ghlatest && ls -al)
