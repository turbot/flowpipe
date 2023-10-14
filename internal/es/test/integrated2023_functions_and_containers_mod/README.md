
## Testing the Lambda function pipeline

Create the credentials file `credentials.auto.pvars`, and add the following:

```sh
aws_access_key_id = "your_access_key_id"
aws_secret_access_key = "your_secret_access_key"
```

Start the Lambda function pipeline:

```sh
make run-test-mod-functions
```

On a separate terminal, run ngrok to expose the Lambda function pipeline:

```sh
ngrok http 7103
```

On another terminal, find the hook URL:

```sh
curl -s http://localhost:7103/api/v0/trigger | jq -r '.items[] | select(.name == "demo.trigger.http.http_trigger_to_iam_policy_validation").url'
```

Build the full hook URL:
  - add the query parameter at the end `?execution_mode=synchronous`
  - combine it with the ngrok URL, plus `api/v0/`

Example:

```sh
https://c130-103-4-88-14.ngrok-free.app/api/v0/hook/test_suite_mod.trigger.http.http_trigger_to_iam_policy_validation/2ab19b09bd3a7d41b920a12eed8e2daf63eb3363b03ab8c1bec0cd2a7d63f833?execution_mode=synchronous
```

Create a HTTPS SNS topic subscription, and add the hook URL as the endpoint

Create a policy with undesired permissions:

```sh
aws iam create-policy --policy-name my-bad-policy1234 --policy-document '{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
                "s3:GetBucketObjectLockConfiguration",
                "s3:DeleteObjectVersion",
                "s3:DeleteBucket"
            ],
            "Resource": "*"
        }
    ]
}'
```
