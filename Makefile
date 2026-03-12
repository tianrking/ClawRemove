APP := claw-remove

.PHONY: test build release

test:
	go test ./...

build:
	mkdir -p dist
	go build -o dist/$(APP) ./cmd/$(APP)

release:
	./scripts/build.sh
