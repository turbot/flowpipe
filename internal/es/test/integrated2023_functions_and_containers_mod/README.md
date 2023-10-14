# Demos
 
### Demo 1 - Lambda Sample App
1. show managed policies (nothing up my sleeve)
2. Create a policy with undesired permissions:

```sh
aws iam create-policy --profile polygoat_c_luis --policy-name my-bad-policy1234 --policy-document '{
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
3.  switch to server console - watch function container images get created and run!
    - note that it installs dependencies!

4. Explain:  
    - This is a sample app (https://aws.amazon.com/blogs/compute/orchestrating-a-security-incident-response-with-aws-step-functions/.).  The lambda code is used as-is!
    - we do the step function logic in flowpipe
    - we added our own function to transform the event data
    - if you update the function code, flowpipe automatically deploys it and creates a new container!
    - currently we support python and nodejs, but we can add support for (most of) the other languages when / if we want. 

4. Open the AWS console, and navigate to IAM -> Policies -> my-bad-policy1234 and confirm that the policy has a second version with the fixed permissions



### Demo 2 - Steampipe / AWS Container 

- explain: will run steampipe check using steampipe container to find alarms for bucket versioning.  Loop through alarms and fix.

- show code (quickly)

- show buckets in alarm:
```bash
steampipe check aws_compliance.benchmark.audit_manager_control_tower_disallow_instances_5_1_1
```

- run the pipeline
    - show split windows: server, client 

```bash
fp pipeline run steampipe_aws --execution-mode synchronous
```

- show fixed:
```bash
steampipe check aws_compliance.benchmark.audit_manager_control_tower_disallow_instances_5_1_1
```


### Demo 3 - Slack DOAWS Container 

- Show first - start stop instance, add tags, list buckets, show users, etc
- Show the code:
    - Trigger, fronted by ngrok, added to the slack app config
    - Pipeline is only 5 steps - 100 lines for the whole thing!
    - Runs http steps and aws-cli container



------
------
------


## Pre-Demo Setup 

### Demo 1 - Lambda Sample App

- Create the credentials file `credentials.auto.pvars`, and add the following:
    ```sh
    aws_access_key_id = "your_access_key_id"
    aws_secret_access_key = "your_secret_access_key"
    ```

- Start the flowpipe server:
    ```sh
    make run-demo
    ```

- On a separate terminal, run ngrok to expose the Lambda function pipeline:
    ```sh
    ngrok http 7103
    ```

- On another terminal, find the hook URL:
    ```bash
    NG_URL=$(curl -s localhost:4040/api/tunnels | jq -r '.tunnels[0].public_url')
    FP_URL=$(curl -s http://localhost:7103/api/v0/trigger | jq -r '.items[] | select(.name == "demo.trigger.http.http_trigger_to_iam_policy_validation").url')
    echo
    echo $NG_URL/api/v0$FP_URL
    ```

- Create a HTTPS SNS topic subscription, and add the hook URL as the endpoin


### Demo 2 - Steampipe / AWS Container 

### Pre-demo Setup
- Build the container 
- set up steampipe (`aws.spc``) with correct account (polygoat_c)
- make sure there are buckets in alarm:
```bash
steampipe check aws_compliance.benchmark.audit_manager_control_tower_disallow_instances_5_1_1
```
- make sure creds are set up in the `flowpipe.pvars` files


### Demo 3 - Slack DOAWS Container 

AWS Account for SNS Topics:  polygoat_c (sso) - #157447638907 | aws+polygoat_c@turbothq.com
- make sure creds are set up in the `flowpipe.pvars` files
- Start flowpipe, start ngrok
- Add webook url ngrok url + trigger url to slack app (Sumit)
    ```bash
    NG_URL=$(curl -s localhost:4040/api/tunnels | jq -r '.tunnels[0].public_url')
    FP_URL=$(curl -s http://localhost:7103/api/v0/trigger | jq -r '.items[] | select(.name == "demo.trigger.http.slack_doaws_webhook_trigger").url')
    echo
    echo $NG_URL/api/v0$FP_URL
    ```
