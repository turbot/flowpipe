
# Functions example

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
