
run:
	FLOWPIPE_LOG_LEVEL=INFO go run . service start --pipeline-dir ./internal/pipeline/collections

run-trace:
	FLOWPIPE_LOG_LEVEL=INFO FLOWPIPE_TRACE_LEVEL=INFO go run . service start --pipeline-dir ./internal/pipeline

clean-tmp:
	rm -rf ./tmp/*

clean-dist:
	rm -rf ./dist/*

clean: clean-tmp clean-dist

build-open-api:
	rm -rf service/api/docs
	./generate-open-api.sh

release-local:
	goreleaser release --snapshot --clean

test:
	go clean -testcache
	go test ./... -timeout 120s -v
