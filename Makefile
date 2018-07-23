USER := marcboudreau
EXECUTABLE := vault-circleci-auth-plugin
RELEASE ?= patch

UNIX_EXECUTABLES := \
    darwin/amd64/$(EXECUTABLE) \
    freebsd/amd64/$(EXECUTABLE) \
    linux/amd64/$(EXECUTABLE) \
    linux/386/$(EXECUTABLE) \
	linux/arm/5/$(EXECUTABLE) \
	linux/arm/7/$(EXECUTABLE)

WINDOWS_EXECUTABLES := \
    windows/amd64/$(EXECUTABLE).exe \
    windows/386/$(EXECUTABLE).exe

COMPRESSED_EXECUTABLES=$(UNIX_EXECUTABLES:%=%.bz2) $(WIN_EXECUTABLES:%.exe=%.zip)
COMPRESSED_EXECUTABLE_TARGETS=$(COMPRESSED_EXECUTABLES:%=bin/%)

all: $(UNIX_EXECUTABLES:%=bin/%) $(WINDOWS_EXECUTABLES:%=bin/%) test-results.txt

# arm
bin/linux/arm/5/$(EXECUTABLE):
	GOARM=5 GOARCH=arm GOOS=linux go build -o "$@"
bin/linux/arm/7/$(EXECUTABLE):
	GOARM=7 GOARCH=arm GOOS=linux go build -o "$@"

# 386
bin/darwin/386/$(EXECUTABLE):
	GOARCH=386 GOOS=darwin go build -o "$@"
bin/linux/386/$(EXECUTABLE):
	GOARCH=386 GOOS=linux go build -o "$@"
bin/windows/386/$(EXECUTABLE).exe:
	GOARCH=386 GOOS=windows go build -o "$@"

# amd64
bin/freebsd/amd64/$(EXECUTABLE):
	GOARCH=amd64 GOOS=freebsd go build -o "$@"
bin/darwin/amd64/$(EXECUTABLE):
	GOARCH=amd64 GOOS=darwin go build -o "$@"
bin/linux/amd64/$(EXECUTABLE):
	GOARCH=amd64 GOOS=linux go build -o "$@"
bin/windows/amd64/$(EXECUTABLE).exe:
	GOARCH=amd64 GOOS=windows go build -o "$@"

# compressed artifacts
%.bz2: %
	bzip2 -c < "$<" > "$@"
%.zip: %.exe
	zip "$@" "$<"

test-results.txt:
	go test -v -race ./... | tee "$@"

tag:
	git semver $(RELEASE)
	
release: all tag
	$(MAKE) $(COMPRESSED_EXECUTABLE_TARGETS)
	git log --format=%B $(shell git semver get) -1 | \
		github-release release -u $(USER) -r $(EXECUTABLE) \
			-t $(shell git semver get) -n $(shell git semver get) -d - || true

clean:
	rm -rf bin/ || true

.PHONY: clean tag release
