
run:
	FLOWPIPE_LOG_LEVEL=INFO go run . service start --pipeline-dir ./internal/es/test/pipelines

run-mod:
	FLOWPIPE_LOG_LEVEL=INFO go run . service start --pipeline-dir ./internal/es/test/default_mod

run-trace:
	FLOWPIPE_LOG_LEVEL=INFO FLOWPIPE_TRACE_LEVEL=INFO go run . service start --pipeline-dir ./internal/es/test/pipelines

clean-tmp:
	rm -rf ./tmp/*

clean-dist:
	rm -rf ./dist/*

clean: clean-tmp clean-dist

build-open-api:
	rm -rf service/api/docs
	./generate-open-api.sh

beta-tag-timetamp:
	date -u +%Y%m%d%H%M

release-local:
	goreleaser release --snapshot --clean

test:
	go clean -testcache
	go test  $$(go list ./... | grep -v /internal/es/test) -timeout 60s -v

integration-test:
	go clean -testcache
	go test ./internal/es/test -timeout 120s -v
