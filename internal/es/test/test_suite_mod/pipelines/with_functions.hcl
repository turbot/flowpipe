pipeline "with_functions" {

    step "function" "hello_nodejs_step" {
        function = function.hello_nodejs
    }

    output "val" {
        value = step.function.hello_nodejs_step.result.body.message
    }

}


pipeline "with_containers" {
    param "aws_region" {
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

    step "container" "hello_nodejs_step" {
        image = "steampipe-aws-compliance"
        cmd = ["steampipe", "check", "aws_compliance.benchmark.audit_manager_control_tower_disallow_instances_5_1_1", "--output", "json"]
        env = {
            AWS_REGION = param.aws_region
            AWS_ACCESS_KEY_ID = param.aws_access_key_id
            AWS_SECRET_ACCESS_KEY = param.aws_secret_access_key
        }
    }
}