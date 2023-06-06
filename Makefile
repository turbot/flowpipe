run:
	FLOWPIPE_LOG_LEVEL=DEBUG go run . service start --pipeline-dir ./pipeline

build-open-api:
	rm -rf service/api/docs
	./generate-open-api.sh

release-local:
	goreleaser release --snapshot --clean
