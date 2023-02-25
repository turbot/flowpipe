
# Functions example

```shell
pushd lambda-python
docker run -p 9000:8080 hello-world-python
popd

pushd lambda-go
docker run -p 9001:8080 hello-world-golang
popd

# Call the python function
curl -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" -d '{}'

# Call the golang function
curl -XPOST "http://localhost:9001/2015-03-31/functions/function/invocations" -d '{"name":"steve2"}'
```
