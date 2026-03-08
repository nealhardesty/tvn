BINARY := tvn
LDFLAGS := -ldflags "-s -w"

.PHONY: build test clean install lint fmt vet push

build:
	go build $(LDFLAGS) -o $(BINARY) .

test:
	go test ./...

test-verbose:
	go test -v ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

lint: fmt vet

clean:
	rm -f $(BINARY)

install: build
	go install $(LDFLAGS) .

# Bump patch version in version.go, commit everything, tag, and push.
push:
	@CURRENT=$$(grep -oE '[0-9]+\.[0-9]+\.[0-9]+' version.go) && \
	NEW=$$(echo $$CURRENT | awk -F. '{printf "%d.%d.%d", $$1, $$2, $$3+1}') && \
	echo "Bumping version: $$CURRENT -> $$NEW" && \
	sed -i "s/\"$$CURRENT\"/\"$$NEW\"/" version.go && \
	git add . && \
	git commit -m "$$(gitsum)" && \
	git push origin && \
	git tag v$$NEW && \
	git push origin v$$NEW && \
	echo "Released v$$NEW"
