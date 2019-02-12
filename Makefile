PKG := github.com/michael1011/lightningtip

GOBUILD := GO111MODULE=on go build -v
GOINSTALL := GO111MODULE=on go install -v

GO_BIN := ${GOPATH}/bin
LINT_BIN := $(GO_BIN)/gometalinter.v2

HAVE_LINTER := $(shell command -v $(LINT_BIN) 2> /dev/null)

default: build

$(LINT_BIN):
	@$(call print, "Fetching gometalinter.v2")
	go get -u gopkg.in/alecthomas/gometalinter.v2

GREEN := "\\033[0;32m"
NC := "\\033[0m"

define print
	echo $(GREEN)$1$(NC)
endef

LINT_LIST = $(shell go list -f '{{.Dir}}' ./...)

LINT = $(LINT_BIN) \
	--disable-all \
	--enable=gofmt \
	--enable=vet \
	--enable=golint \
	--line-length=72 \
	--deadline=4m $(LINT_LIST) 2>&1 | \
	grep -v 'ALL_CAPS\|OP_' 2>&1 | \
	tee /dev/stderr

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
	GO111MODULE=on go mod vendor
	GO111MODULE=off $(LINT_BIN) --install 1> /dev/null
	test -z "$$($(LINT))"
