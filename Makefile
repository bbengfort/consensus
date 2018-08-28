# Scripts to handle consensus build and installation
# Shell to use with Make
SHELL := /bin/bash

# Build Environment
PACKAGE = consensus
PBPKG = $(CURDIR)/pb
BUILD = $(CURDIR)/_build

# Commands
GOCMD = go
GODEP = dep ensure
GODOC = godoc
GINKGO = ginkgo
PROTOC = protoc
GORUN = $(GOCMD) run
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean

# Output Helpers
BM  = $(shell printf "\033[34;1m●\033[0m")
GM = $(shell printf "\033[32;1m●\033[0m")
RM = $(shell printf "\033[31;1m●\033[0m")


# Export targets not associated with files.
.PHONY: all install build consensus test citest clean doc protobuf

# Ensure dependencies are installed, run tests and compile
all: test build

# Compile protocol buffers
protobuf:
	$(info $(GM) compiling protocol buffers …)
	@ $(PROTOC) -I $(PBPKG) $(PBPKG)/*.proto --go_out=plugins=grpc:$(PBPKG)

# Install the commands and create configurations and data directories
install: build
	$(info $(GM) installing consensus and making configuration …)
	@ cp $(BUILD)/consensus /usr/local/bin/

# Build the various binaries and sources
build: protobuf consensus

# Build the consensus command and store in the build directory
consensus:
	$(info $(GM) compiling consensus executable …)
	@ $(GOBUILD) -o $(BUILD)/consensus ./cmd/consensus

# Target for simple testing on the command line
test:
	$(info $(BM) running simple local tests …)
	@ $(GINKGO) -r

# Target for testing in continuous integration
citest:
	$(info $(BM) running CI tests with randomization and race …)
	$(GINKGO) -r -v --randomizeAllSpecs --randomizeSuites --failOnPending --cover --trace --race --compilers=2

# Run Godoc server and open browser to the documentation
doc:
	$(info $(BM) running go documentation server at http://localhost:6060)
	$(info $(BM) type CTRL+C to exit the server)
	@ open http://localhost:6060/pkg/github.com/bbengfort/consensus/
	@ $(GODOC) --http=:6060

# Clean build files
clean:
	$(info $(RM) cleaning up build …)
	@ $(GOCLEAN)
	@ find . -name "*.coverprofile" -print0 | xargs -0 rm -rf
	@ rm -rf $(BUILD)
