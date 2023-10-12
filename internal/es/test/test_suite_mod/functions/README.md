
## Testing the Lambda function pipeline

One one terminal, run the following command to start the Lambda function pipeline:

```sh
make run-test-mod-functions
```

On another terminal, run the following command to test it:

```sh
curl --location 'http://localhost:7103/api/v0/pipeline/lambda_example/cmd' \
--header 'Content-Type: application/json' \
--data '{
    "command": "run",
    "execution_mode": "synchronous",
    "args": {
        "event": {
            "policy": {
                "Version":"2012-10-17","Statement":[{"Sid":"VisualEditor0","Effect":"Allow","Action":["s3:DeleteObject"],"Resource":"*"}]
            },
            "policyMeta": {
                "arn": "arn:aws:iam::123456789012:policy/ExamplePolicy",
                "policyName": "ExamplePolicy",
                "defaultVersionId": "v1"
            }
        }
    }
}'
```
