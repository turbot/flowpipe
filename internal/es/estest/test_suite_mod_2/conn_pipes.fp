pipeline "pipe_with_conn_param" {

    param "region" {
        type        = string
        description = "The name of the region."
        default     = "ap-southeast-2"
    }

    param "conn" {
        type        = connection.aws
        description = "AWS connection to connect with"
        #default     = var.conn
        default     = connection.aws.default
    }

    step "container" "describe_vpcs" {
        image = "amazon/aws-cli"

        cmd = concat(
        ["ec2", "describe-vpcs"],
        try(length(param.vpc_ids), 0) > 0 ? concat(["--vpc-ids"], param.vpc_ids) : []
        )

        # this works:
        # env = merge(connection.aws.default.env, {AWS_REGION = param.region})

        # but this doesnt - env{} is empty
        env = merge(param.conn.env, {AWS_REGION = param.region})

    }

    step "sleep" "sleep" {
        depends_on = [step.container.describe_vpcs]
        duration = "60s"
    }

  output "stdout" {
    description = "The standard output stream from the AWS CLI."
    value       = jsondecode(step.container.describe_vpcs.stdout)
  }

  output "stderr" {
    description = "The standard error stream from the AWS CLI."
    value       = step.container.describe_vpcs.stderr
  }

}