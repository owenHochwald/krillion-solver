GO ?= go
DIST ?= dist
XCODEBUILD ?= xcodebuild
BINARY := krillion$(if $(filter Windows_NT,$(OS)),.exe,)
MENU_BAR_APP := $(DIST)/KrillionBar.app
MENU_BAR_DERIVED := $(DIST)/xcode

.PHONY: build web macos macos-app menu-bar menu-bar-build linux windows release test vet clean

build:
	$(GO) build -o $(BINARY) .

web:
	$(GO) run . --web

macos:
	@mkdir -p $(DIST)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GO) build -o $(DIST)/krillion-darwin-arm64 .
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GO) build -o $(DIST)/krillion-darwin-amd64 .

menu-bar-build:
	$(XCODEBUILD) -quiet -project macos/KrillionBar.xcodeproj -scheme KrillionBar -configuration Release -derivedDataPath $(MENU_BAR_DERIVED) CODE_SIGNING_ALLOWED=NO build
	@rm -rf -- "$(MENU_BAR_APP)"
	@ditto "$(MENU_BAR_DERIVED)/Build/Products/Release/KrillionBar.app" "$(MENU_BAR_APP)"
	@printf 'Built %s\n' "$(MENU_BAR_APP)"

menu-bar: menu-bar-build
	open "$(MENU_BAR_APP)"

macos-app: menu-bar-build

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
