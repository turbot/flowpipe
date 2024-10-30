pipeline "create_s3_bucket" {
  title       = "Create S3 Bucket"
  description = "Creates a new Amazon S3 bucket."

  param "aws_region" {
    type = string
    default = "us-east-1"
  }
  param "aws_access_key_id" {
    type = string
    default = "abc"
  }
  param "aws_secret_access_key" {
    type = string
    default = "abc"
  }

  param "bucket" {
    type        = string
    description = "The name of the new S3 bucket."
    default     = "test-bucket"
  }

  param "acl" {
    type        = string
    description = "The access control list (ACL) for the new bucket (e.g., private, public-read)."
    optional = true
  }

  step "container" "create_s3_bucket" {
    image = "amazon/aws-cli"

    cmd = concat(
      ["s3api", "create-bucket_xxxxx"],
      ["s3api", "create-bucket_xxxxx"],
    )

    env = {
      AWS_REGION            = param.aws_access_key_id,
      AWS_ACCESS_KEY_ID     = param.aws_access_key_id,
      AWS_SECRET_ACCESS_KEY = param.aws_secret_access_key
    }
  }

  output "stdout" {
    description = "The JSON output from the AWS CLI."
    value       = step.container.create_s3_bucket.stdout
  }

  output "stderr" {
    description = "The error output from the AWS CLI."
    value       = step.container.create_s3_bucket.stderr
  }
}