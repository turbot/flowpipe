pipeline "with_bad_container" {
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

    step "container" "container_run" {
        image = "amazon/aws-cli"

        cmd = ["sts", "get-caller-identity"]

        env = {
            AWS_REGION = param.aws_region
            AWS_ACCESS_KEY_ID = param.aws_access_key_id
            AWS_SECRET_ACCESS_KEY = param.aws_secret_access_key
        }
    }
}

pipeline "with_bad_container_with_is_error" {
    step "pipeline" "create_s3_bucket" {
        pipeline = pipeline.create_s3_bucket

        error {
            ignore = true
        }
    }

    step "pipeline" "delete_s3_bucket" {
        if = !is_error(step.pipeline.create_s3_bucket)

        pipeline = pipeline.delete_s3_bucket

        error {
            ignore = true
        }
    }
}

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
      ["s3api", "create-bucket"],
      ["--bucket", param.bucket],
      param.acl != null ? ["--acl", param.acl] : [],
      # Regions other than us-east-1 require the LocationConstraint parameter
      param.aws_region != "us-east-1" ? ["--create-bucket-configuration LocationConstraint=", param.aws_region] : [],
    )

    env = {
      AWS_REGION            = param.aws_region,
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
pipeline "delete_s3_bucket" {
  title       = "Delete S3 Bucket"
  description = "Deletes an Amazon S3 bucket."

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

  step "container" "delete_s3_bucket" {
    image = "amazon/aws-cli"

    cmd = [
      "s3api",
      "delete-bucket",
      "--bucket", param.bucket
    ]

    env = {
      AWS_REGION            = param.aws_region,
      AWS_ACCESS_KEY_ID     = param.aws_access_key_id,
      AWS_SECRET_ACCESS_KEY = param.aws_secret_access_key
    }
  }

  output "stdout" {
    description = "The JSON output from the AWS CLI."
    value       = step.container.delete_s3_bucket.stdout
  }

  output "stderr" {
    description = "The error output from the AWS CLI."
    value       = step.container.delete_s3_bucket.stderr
  }
}

