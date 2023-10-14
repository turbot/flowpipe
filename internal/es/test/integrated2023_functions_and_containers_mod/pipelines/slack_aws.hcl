trigger "http" "http_trigger_to_disable_versioning" {
    title    = "Webhook Trigger for Slack /fix command"
    pipeline = pipeline.disable_versioning
    args     = {
      response_url = parse_query_string(self.request_body).response_url
      bucket_name = parse_query_string(self.request_body).text
    }
}

pipeline "disable_versioning" {
    param "response_url" {
        description = "The url to respond to slack"
        type        = string
    }

    param "bucket_name" {
        description = "The bucket name that the user passed to the /fix command" 
        type        = string
    }

    step "http" "on_it" {
        description = "Echo the user's request back to them"
        url         = param.response_url
        method      = "post"

        request_body = jsonencode({
            text = "Disabling ${param.bucket_name}..."
        })
    }

    step "container" "container_run_aws" {
        description = "Run the AWS cli command in the aws-cli container"
        image       = "amazon/aws-cli"
        cmd         = ["s3api", "put-bucket-versioning", "--bucket", param.bucket_name, "--versioning-configuration", "Status=Suspended"]
        
        env = {
            AWS_REGION = var.aws_region
            AWS_ACCESS_KEY_ID = var.aws_access_key_id
            AWS_SECRET_ACCESS_KEY = var.aws_secret_access_key
        }
    }

    step "http" "done" {
        description = "Respond to slack with the acknowledgement"
        depends_on  = [step.container.container_run_aws]
        url = param.response_url
        method = "post"
        request_body = jsonencode({
            text = "Done"
        })
    }

    output "stdout_aws" {
        description = "The aws-cli command output"
        value       = step.container.container_run_aws.stdout
    }

     output "stderr_aws" {
        description = "The aws-cli command error"
        value       = step.container.container_run_aws.stderr
    }
}

