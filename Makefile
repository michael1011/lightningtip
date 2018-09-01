PKG := github.com/michael1011/lightningtip

GOBUILD := go build -v
GOINSTALL := go install -v

GO_BIN := ${GOPATH}/bin
DEP_BIN := $(GO_BIN)/dep
LINT_BIN := $(GO_BIN)/gometalinter.v2

HAVE_DEP := $(shell command -v $(DEP_BIN) 2> /dev/null)
HAVE_LINTER := $(shell command -v $(LINT_BIN) 2> /dev/null)

$(LINT_BIN):
	@$(call print, "Fetching gometalinter.v2")
	go get -u gopkg.in/alecthomas/gometalinter.v2

$(DEP_BIN):
	@$(call print, "Fetching dep")
	go get -u github.com/golang/dep/cmd/dep

LINT_LIST = $(go list -f '{{.Dir}}' ./...)

GREEN := "\\033[0;32m"
NC := "\\033[0m"

define print
	echo $(GREEN)$1$(NC)
endef

LINT = $(LINT_BIN) $(LINT_LIST) \
	--disable-all \
	--enable=gofmt \
	--enable=vet \
	--enable=golint \
	--line-length=72 \
	--deadline=4m $(GOLISTLINT) 2>&1 | \
	grep -v 'ALL_CAPS\|OP_' 2>&1 | \
	tee /dev/stderr

default: dep build

# Dependencies

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

# Utils

fmt:
	@$(call print, "Formatting source")
	gofmt -s -w .

lint: $(LINT_BIN)
	@$(call print, "Linting source")
	$(LINT_BIN) --install 1> /dev/null
	test -z "$$($(LINT))"
