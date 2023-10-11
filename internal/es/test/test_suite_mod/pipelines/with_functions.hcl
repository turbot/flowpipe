pipeline "with_functions" {

    step "function" "hello_nodejs_step" {
        function = function.hello_nodejs
    }

    output "val" {
        value = step.function.hello_nodejs_step.result.body.message
    }

}


pipeline "with_container" {
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
        image = "steampipe-aws-compliance"
        cmd = ["steampipe", "check", "aws_compliance.benchmark.audit_manager_control_tower_disallow_instances_5_1_1", "--output", "json"]
        env = {
            AWS_REGION = param.aws_region
            AWS_ACCESS_KEY_ID = param.aws_access_key_id
            AWS_SECRET_ACCESS_KEY = param.aws_secret_access_key
        }
    }

    output "stdout" {
        value = step.container.container_run.stdout
    }

    output "stderr" {
        value = step.container.container_run.stderr
    }
}