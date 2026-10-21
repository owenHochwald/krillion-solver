GO ?= go
DIST ?= dist
BINARY := krillion$(if $(filter Windows_NT,$(OS)),.exe,)

.PHONY: build macos macos-app linux windows release test vet clean

build:
	$(GO) build -o $(BINARY) .

macos:
	@mkdir -p $(DIST)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GO) build -o $(DIST)/krillion-darwin-arm64 .
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GO) build -o $(DIST)/krillion-darwin-amd64 .

macos-app:
	xcodebuild -project macos/KrillionBar.xcodeproj -scheme KrillionBar -configuration Release -derivedDataPath $(DIST)/xcode CODE_SIGNING_ALLOWED=NO build

linux:
	@mkdir -p $(DIST)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GO) build -o $(DIST)/krillion-linux-arm64 .
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build -o $(DIST)/krillion-linux-amd64 .

windows:
	@mkdir -p $(DIST)
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 $(GO) build -o $(DIST)/krillion-windows-arm64.exe .
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GO) build -o $(DIST)/krillion-windows-amd64.exe .

release: macos linux windows

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

clean:
	rm -rf $(DIST) krillion krillion.exe
