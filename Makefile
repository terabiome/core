.PHONY: clone-goose build-goose goose-binary

SHELL=/bin/bash
GOOSE_RELEASE_TAG=v3.29.0
GOOSE_GIT_PATH=/tmp/goose-migration-${GOOSE_RELEASE_TAG}
GOOSE_BIN_PATH=$(shell go env GOBIN)/goose
TRAILING_ARGUMENTS=$(wordlist 2, $(words $(MAKECMDGOALS)), $(MAKECMDGOALS))

clone-goose:
	@git clone \
		--branch ${GOOSE_RELEASE_TAG} \
		--single-branch \
		https://github.com/pressly/goose ${GOOSE_GIT_PATH} \
	&& echo "Downloaded Goose repository" || echo "Failed to download Goose repository"

build-goose:
	@cd ${GOOSE_GIT_PATH} \
	&& go build -mod=mod -tags='\
		no_azuresql \
		no_clickhouse \
		no_libsql \
		no_mssql \
		no_mysql \
		no_vertica \
		no_ydb \
		' -o ${GOOSE_BIN_PATH} ${GOOSE_GIT_PATH}/cmd/goose \
	&& echo "Built Goose binary"

adopt-goose: clone-goose build-goose

clean:
	@rm -rf ${GOOSE_GIT_PATH} && echo "Removed Goose repository"
	@rm -f ${GOOSE_BIN_PATH} && echo "Remove Goose binary"
