PACKAGE_NAME          := github.com/turbot/flowpipe
GOLANG_CROSS_VERSION  ?= v1.21.5

.PHONY: build
build: build-ui
	go build .

run-mod:
	go run . server --mod-location ./internal/es/estest/default_mod

run-test-mod:
	FP_VAR_var_from_env="from env var" go run . server --mod-location ./internal/es/estest/test_suite_mod

run-pipeline:
	FLOWPIPE_LOG_LEVEL=INFO go run . server --mod-location ./internal/es/estest/pipelines

run-trace:
	FLOWPIPE_LOG_LEVEL=INFO FLOWPIPE_TRACE_LEVEL=INFO go run . server --mod-location ./internal/es/estest/pipelines

clean-tmp:
	rm -rf ./internal/es/estest/test_suite_mod/.flowpipe/store/*
	rm -f ./internal/es/estest/test_suite_mod/flowpipe.db
	rm -f flowpipe

clean-dist:
	rm -rf ./dist/*

clean-debug:
	rm -rf __debug*

clean: clean-tmp clean-dist clean-debug

build-open-api:
	rm -rf service/api/docs && cp -f ./scripts/generate-open-api.sh . && ./generate-open-api.sh

beta-tag-timetamp:
	date -u +%Y%m%d%H%M

.PHONY: build-ui
build-ui:
	cd ui/form && yarn install && yarn build

.PHONY: test
test:
	go clean -testcache
	RUN_MODE=TEST_ES go test  $$(go list ./... | grep -v /internal/es/estest) -timeout 60s

.PHONY: integration-test
integration-test:
	go clean -testcache
	RUN_MODE=TEST_ES go test ./internal/es/estest -timeout 240s -v

.PHONY: release-dry-run
release-dry-run:
	@docker run \
		--rm \
		-e CGO_ENABLED=1 \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/flowpipe \
		-v `pwd`/../pipe-fittings:/go/src/pipe-fittings \
		-v `pwd`/../flowpipe-sdk-go:/go/src/flowpipe-sdk-go \
		-w /go/src/flowpipe \
		ghcr.io/goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
		--clean --skip=validate --skip=publish


.PHONY: release
release:
	@if [ ! -f ".release-env" ]; then \
		echo ".release-env is required for release";\
		exit 1;\
	fi
	docker run \
		--rm \
		-e CGO_ENABLED=1 \
		--env-file .release-env \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/flowpipe \
		-v `pwd`/../pipe-fittings:/go/src/pipe-fittings \
		-v `pwd`/../flowpipe-sdk-go:/go/src/flowpipe-sdk-go \
		-w /go/src/flowpipe \
		ghcr.io/goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
		release --clean
