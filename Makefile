CGO_ENABLED=0
GOOS=linux
GOARCH=amd64
COMMIT=`git rev-parse --short HEAD`
APP=finca
ORG=ehazlett
REPO?=$(ORG)/$(APP)
TAG?=latest
DEPS=$(shell go list ./... | grep -v /vendor/)

all: build

build:
	@cd cmd/$(APP) && go build -v -a -tags "netgo static_build" -installsuffix netgo -ldflags "-w -X github.com/$(REPO)/version.GitCommit=$(COMMIT)" .

image: build
	@docker build -t $(REPO):$(TAG) .

release: image
	@docker push $(REPO):$(TAG)

check:
	@go vet -v $(DEPS)
	@golint $(DEPS)


test: build
	@go test -v ./...

clean:
	@rm -rf cmd/$(APP)/$(APP)
	@rm -rf build

.PHONY: build image release test clean
