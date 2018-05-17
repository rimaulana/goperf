BINARY := goperf
TEST_MODE ?= count
VERSION ?= latest
BIN_DIR := $(GOPATH)/bin 
GOLINT = $(BIN_DIR)/golint
PACKAGES = $(shell go list ./... | grep -v vendor)
GOLINT_REPO = github.com/golang/lint/golint
PLATFORM := windows linux
OS = $(word 1, $@)

.PHONY: lint test clean cover

$(GOLINT):
	go get -u $(GOLINT_REPO)

lint: $(GOLINT)
	for PKG in $(PACKAGES); do \
		golint -set_exit_status $$PKG || exit 1; \
	done;

clean:
	rm -f coverage.html; \
	rm -f c.out; \
	rm -rf release;

test: clean
	echo "mode: $(TEST_MODE)" > c.out; \
	for PKG in $(PACKAGES); do \
		go test -v -covermode=$(TEST_MODE) -coverprofile=profile.out $$PKG; \
		if [ -f profile.out ]; then \
        	cat profile.out | grep -v "mode:" >> c.out; \
        	rm profile.out; \
    	fi; \
	done;

cover: test
	go tool cover -html=c.out -o=coverage.html; \
	rm -f c.out;