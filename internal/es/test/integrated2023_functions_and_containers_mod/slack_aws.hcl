pipeline "disable_versioning" {
    param "response_url" {
        type = string
    }
    param "bucket_name" {
        type = string
    }
    step "http" "on_it" {
        url = param.response_url
        method = "post"
        request_body = jsonencode({
            text = "Disabling ${param.bucket_name}..."
        })
    }
    step "container" "container_run_aws" {
        image = "amazon/aws-cli"
        cmd = ["s3api", "put-bucket-versioning", "--bucket", param.bucket_name, "--versioning-configuration", "Status=Suspended"]
        env = {
            AWS_REGION = var.aws_region
            AWS_ACCESS_KEY_ID = var.aws_access_key_id
            AWS_SECRET_ACCESS_KEY = var.aws_secret_access_key
        }
    }
    step "http" "done" {
        depends_on = [step.container.container_run_aws]
        url = param.response_url
        method = "post"
        request_body = jsonencode({
            text = "Done"
        })
    }
    output "stdout_aws" {
        value = step.container.container_run_aws.stdout
    }
     output "stderr_aws" {
        value = step.container.container_run_aws.stderr
    }
    output "response_url" {
        value = param.response_url
    }
}

trigger "http" "http_trigger_to_disable_versioning" {

    pipeline = pipeline.disable_versioning
    args     = {
      response_url = parse_query_string(self.request_body).response_url
      bucket_name = parse_query_string(self.request_body).text
    }
}


