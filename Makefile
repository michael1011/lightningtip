PKG := github.com/michael1011/lightningtip

GOBUILD := go build -v
GOINSTALL := go install -v

GO_BIN := ${GOPATH}/bin
DEP_BIN := $(GO_BIN)/dep

HAVE_DEP := $(shell command -v $(DEP_BIN) 2> /dev/null)

GREEN := "\\033[0;32m"
NC := "\\033[0m"

define print
	echo $(GREEN)$1$(NC)
endef

default: scratch

# Dependencies

$(DEP_BIN):
	@$(call print, "Fetching dep")
	go get -u github.com/golang/dep/cmd/dep

dep: $(DEP_BIN)
	@$(call print, "Compiling dependencies")
	dep ensure -v

# Building

build:
	@$(call print, "Building lightningtip and tipreport")
	$(GOBUILD) -o lightningtip $(PKG)
	$(GOBUILD) -o tipreport $(PKG)/cmd/tipreport

install:
	@$(call print, "Installing lightningtip and tipreport")
	$(GOINSTALL) $(PKG)
	$(GOINSTALL) $(PKG)/cmd/tipreport

scratch: dep build