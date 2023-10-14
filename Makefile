
run-mod:
	FLOWPIPE_LOG_LEVEL=INFO go run . service start --mod-location ./internal/es/test/default_mod

run-test-mod:
	P_VAR_var_from_env="from env var" FLOWPIPE_LOG_LEVEL=INFO go run . service start --mod-location ./internal/es/test/test_suite_mod --log-dir ./tmp --output-dir ./tmp

run-test-mod-functions:
	P_VAR_var_from_env="from env var" FLOWPIPE_LOG_LEVEL=INFO go run . service start --mod-location ./internal/es/test/test_suite_mod --functions --log-dir ./tmp --output-dir ./tmp

run-demo:
	FLOWPIPE_LOG_LEVEL=INFO go run . service start --mod-location ./internal/es/test/integrated2023_functions_and_containers_mod --functions --log-dir ./tmp --output-dir ./tmp

run-pipeline:
	FLOWPIPE_LOG_LEVEL=INFO go run . service start --mod-location ./internal/es/test/pipelines

run-trace:
	FLOWPIPE_LOG_LEVEL=INFO FLOWPIPE_TRACE_LEVEL=INFO go run . service start --mod-location ./internal/es/test/pipelines

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
	RUN_MODE=TEST_ES go test  $$(go list ./... | grep -v /internal/es/test) -timeout 60s -v

integration-test:
	go clean -testcache
	RUN_MODE=TEST_ES go test ./internal/es/test -timeout 120s -v
