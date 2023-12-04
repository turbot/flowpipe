run-mod:
	FLOWPIPE_LOG_LEVEL=INFO go run . server --mod-location ./internal/es/estest/default_mod

run-test-mod:
	FP_VAR_var_from_env="from env var" FLOWPIPE_LOG_LEVEL=INFO go run . server --mod-location ./internal/es/estest/test_suite_mod --log-dir ./tmp --output-dir ./tmp

run-test-mod-functions:
	FP_VAR_var_from_env="from env var" FLOWPIPE_LOG_LEVEL=INFO go run . server --mod-location ./internal/es/estest/test_suite_mod --functions --log-dir ./tmp --output-dir ./tmp

run-pipeline:
	FLOWPIPE_LOG_LEVEL=INFO go run . server --mod-location ./internal/es/estest/pipelines

run-trace:
	FLOWPIPE_LOG_LEVEL=INFO FLOWPIPE_TRACE_LEVEL=INFO go run . server --mod-location ./internal/es/estest/pipelines

clean-tmp:
	rm -rf ./internal/es/estest/test_suite_mod/.flowpipe/store/*

clean-dist:
	rm -rf ./dist/*

clean: clean-tmp clean-dist

build-open-api:
	rm -rf service/api/docs
	./generate-open-api.sh

beta-tag-timetamp:
	date -u +%Y%m%d%H%M

release-local:
	# --snapshot means that Go Releaser will not check if the repo is dirty to use the latest atg
	# it simply use the latest commit for the build
	goreleaser release --snapshot --clean

test:
	go clean -testcache
	RUN_MODE=TEST_ES go test  $$(go list ./... | grep -v /internal/es/estest) -timeout 60s -v

integration-test:
	go clean -testcache
	RUN_MODE=TEST_ES go test ./internal/es/estest -timeout 120s -v
