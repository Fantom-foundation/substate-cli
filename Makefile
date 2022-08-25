.PHONY: all
all: substate-cli super-instruction

GOPROXY ?= "https://proxy.golang.org,direct"
.PHONY: substate-cli super-instruction

substate-cli:
	GIT_COMMIT=`git rev-list -1 HEAD 2>/dev/null || echo ""` && \
	GIT_DATE=`git log -1 --date=short --pretty=format:%ct 2>/dev/null || echo ""` && \
	GOPROXY=$(GOPROXY) \
	go build \
	    -ldflags "-s -w -X github.com/Fantom-foundation/substate-cli/cmd/substate-cli/replay.gitCommit=$${GIT_COMMIT} -X github.com/Fantom-foundation/substate-cli/cmd/substate-cli/replay.gitDate=$${GIT_DATE}" \
	    -o build/substate-cli \
	    ./cmd/substate-cli

super-instruction:
	GIT_COMMIT=`git rev-list -1 HEAD 2>/dev/null || echo ""` && \
	GIT_DATE=`git log -1 --date=short --pretty=format:%ct 2>/dev/null || echo ""` && \
	GOPROXY=$(GOPROXY) \
	go build \
	    -ldflags "-s -w" \
	    -o build/super-instruction \
	    ./cmd/super-instruction


.PHONY: clean
clean:
	rm -fr ./build/*
