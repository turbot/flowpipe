
# Functions example

## Usage

```shell
pushd lambda-python
docker build -t hello-world-python .
docker run -p 9000:8080 hello-world-python
popd

pushd lambda-go
docker build -t hello-world-golang .
docker run -p 9001:8080 hello-world-golang
popd

pushd lambda-nodejs
docker build -t hello-world-nodejs .
docker run -p 9002:8080 hello-world-nodejs
popd

# Call the python function
curl -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" -d '{}'

# Call the golang function
curl -XPOST "http://localhost:9001/2015-03-31/functions/function/invocations" -d '{"name":"steve2"}'

# Call the golang function
curl -XPOST "http://localhost:9002/2015-03-31/functions/function/invocations" -d '{"name":"steve2"}'
```

## Reading

Great:
* https://www.bschaatsbergen.com/behind-the-scenes-lambda/
* https://www.sosnowski.dev/post/anatomy-of-aws-lambda#what-s-under-the-hood-

Also helpful:
* https://www.dynatrace.com/news/blog/a-look-behind-the-scenes-of-aws-lambda-and-our-new-lambda-monitoring-extension/

## Architecture

Services:
* Build service - watches code, builds images, ensures available in (local) registry
* Run service - runs containers and handles routing by version
* Hooks service - webhook listeners, communicating with the run service

Dump of notes:
* Docker must be running - detect nicely when it's not
* Use AWS Lambda docker containers
* Watch for code changes
* Each code change should be an overall "version", all functions should work in that version together. Simulating the idea of a pipeline run using consistent versions.
* Set standard labels on built images - https://github.com/opencontainers/image-spec/blob/main/annotations.md
* Need a tagging scheme for images
* Each function is a microservice, called via HTTP
* Every function should have it's own image
* Every function version should have it's own tag within the image
* Q - Can a given microservice be called multiple times in parallel? Or do we need separate containers for each invocation?
* Existing requests should use their version to be served, new requests are sent to the new version
* Starting a new function version should be (temporarily) in parallel to the existing version
* Should support limits on concurrency (e.g. min and max)
* Should support limits on memory and CPU usage / allocation
* Need to pass in input and get output
* Need to be able to pass in environment variables
* Need to ensure base image is patched on reasonable schedule (docker cache can cause it to get backlogged)
* Validation of function names (i.e. image names)