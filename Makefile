.GIT_COMMIT=$(shell git rev-parse HEAD)
.GIT_COMMIT_DATE=$(shell git log -1 --date=format:'%Y-%m-%dT%H:%M:%S' --format=%cd)
.GIT_VERSION=$(shell git describe --tags 2>/dev/null || echo "$(.GIT_COMMIT)")
.LD_FLAGS=$(shell echo "-s -w -X main.appVersion=${.GIT_VERSION} -X main.appCommit=${.GIT_COMMIT} -X main.appCommitDate=${.GIT_COMMIT_DATE}")

all: deps test vet fmt lint install

deps:
	go install github.com/mgechev/revive@v1.3.3
	go mod tidy

test:
	go test ./...

int_test:
	go test -tags=int_test ./...

vet:
	go vet ./...

fmt:
	gofmt -l -w .

lint:
	revive -config revive.toml -formatter friendly ./...

build:
	go build --ldflags "${.LD_FLAGS}" -o bin/aem ./cmd/aem

install:
	go install --ldflags "${.LD_FLAGS}" ./cmd/aem

other_build:
	GOARCH=amd64 GOOS=darwin go build --ldflags "${.LD_FLAGS}" -o bin/aem.darwin ./cmd/aem
	GOARCH=amd64 GOOS=linux go build --ldflags "${.LD_FLAGS}" -o bin/aem.linux ./cmd/aem
	GOARCH=amd64 GOOS=windows go build -tags timetzdata --ldflags "${.LD_FLAGS}" -o bin/aem.exe ./cmd/aem

clean:
	go clean
	rm -fr bin
