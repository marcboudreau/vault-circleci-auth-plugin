USER := marcboudreau
EXECUTABLE := vault-circleci-auth-plugin

UNIX_EXECUTABLES := \
    darwin/amd64/$(EXECUTABLE) \
    freebsd/amd64/$(EXECUTABLE) \
    linux/amd64/$(EXECUTABLE) \
    linux/386/$(EXECUTABLE)

WINDOWS_EXECUTABLES := \
    windows/amd64/$(EXECUTABLE) \
    windows/386/$(EXECUTABLE)

COMPRESSED_EXECUTABLES=$(UNIX_EXECUTABLES:%=%.bz2) $(WIN_EXECUTABLES:%.exe=%.zip)
COMPRESSED_EXECUTABLE_TARGETS=$(COMPRESSED_EXECUTABLES:%=bin/%)

all: $(EXECUTABLE)

# the executable used to perform the upload, dogfooding and all...
bin/tmp/$(EXECUTABLE):
	go build -o "$@"

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
bin/windows/386/$(EXECUTABLE):
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

$(EXECUTABLE):
	go build -o "$@"
	go test -v -race ./...

tag:
	VERSION=$(gen-version.sh $(RELEASE_TYPE))
	github-tag.sh $(VERSION)
	LAST_TAG=v$(VERSION)

release: clean tag
	$(MAKE) $(COMPRESSED_EXECUTABLE_TARGETS)
	git log --format=%B $(LAST_TAG) -1 | \
		docker run release -u $(USER) -r $(EXECUTABLE) \
			-t $(LAST_TAG) -n $(LAST_TAG) -d - || true

clean:
	rm go-app || true
	rm $(EXECUTABLE) || true
	rm -rf bin/

.PHONY: clean release
