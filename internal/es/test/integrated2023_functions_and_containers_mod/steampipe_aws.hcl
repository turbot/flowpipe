pipeline "steampipe_aws" {

    step "container" "container_run_steampipe" {
        description = "Run the Steampipe check cli command in the steampipe-aws-compliance container"
        image       = "steampipe-aws-compliance"
        cmd         = ["steampipe", "check", "aws_compliance.benchmark.audit_manager_control_tower_disallow_instances_5_1_1", "--output", "json"]
        env = {
            AWS_REGION = var.aws_region
            AWS_ACCESS_KEY_ID = var.aws_access_key_id
            AWS_SECRET_ACCESS_KEY = var.aws_secret_access_key
        }
    }

    step "container" "container_run_aws" {
        description = "Run the AWS cli command in the aws-cli container"
        for_each    = jsondecode(step.container.container_run_steampipe.stdout).groups[0].controls[0].results
        if          = each.value.status == "alarm"
        image       = "amazon/aws-cli"
        cmd         = ["s3api", "put-bucket-versioning", "--bucket", element(split(":", each.value.resource), 5), "--versioning-configuration", "Status=Enabled"]
        env = {
            AWS_REGION = var.aws_region
            AWS_ACCESS_KEY_ID = var.aws_access_key_id
            AWS_SECRET_ACCESS_KEY = var.aws_secret_access_key
        }
    }

    output "stdout_aws" {
        value = values(step.container.container_run_aws)[*].stdout
    }

}