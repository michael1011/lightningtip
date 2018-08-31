# TODO: lint and fmt sub packages
PKG := github.com/michael1011/lightningtip

GOBUILD := go build -v
GOINSTALL := go install -v

GO_BIN := ${GOPATH}/bin
DEP_BIN := $(GO_BIN)/dep
LINT_BIN := $(GO_BIN)/gometalinter.v2

HAVE_DEP := $(shell command -v $(DEP_BIN) 2> /dev/null)

GOLIST := go list $(PKG)/... | grep -v '/vendor/'

GREEN := "\\033[0;32m"
NC := "\\033[0m"

define print
	echo $(GREEN)$1$(NC)
endef

LINT = $(LINT_BIN) \
	--disable-all \
	--enable=gofmt \
	--enable=vet \
	--enable=golint \
	--line-length=72 \
	--deadline=4m $(GOLISTLINT) 2>&1 | \
	grep -v 'ALL_CAPS\|OP_' 2>&1 | \
	tee /dev/stderr

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

# Utils

fmt:
	@$(call print, "Formatting source.")
	$(GOLIST) | go fmt -x

lint: $(LINT_BIN)
	@$(call print, "Linting source.")
	$(LINT_BIN) --install 1> /dev/null
	test -z "$$($(LINT))"
