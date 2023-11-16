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

        cmd = concat(
          ["aws", "sts"],
          ["get-caller-identity"]
        )

        env = {
            AWS_REGION = param.aws_region
            AWS_ACCESS_KEY_ID = param.aws_access_key_id
            AWS_SECRET_ACCESS_KEY = param.aws_secret_access_key
        }
    }

    # output "stdout" {
    #     value = step.container.container_run.stdout
    # }

    # output "stderr" {
    #     value = step.container.container_run.stderr
    # }

    # output "combined" {
    #     value = step.container.container_run.combined
    # }
}