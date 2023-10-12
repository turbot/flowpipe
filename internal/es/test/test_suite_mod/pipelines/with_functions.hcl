pipeline "with_functions" {

    param "event" {
        type = any
    }

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

    step "function" "hello_nodejs_step" {
        runtime = "nodejs:18"
        handler = "index.handler"
        src = "./functions/hello-nodejs"
        event = param.event

        env = {
            AWS_REGION = param.aws_region
            AWS_ACCESS_KEY_ID = param.aws_access_key_id
            AWS_SECRET_ACCESS_KEY = param.aws_secret_access_key
            FOO = "bar"
        }
    }

    output "val" {
        value = step.function.hello_nodejs_step.result.body.message
    }

    output "env" {
        value = step.function.hello_nodejs_step.result.body.env
    }

    output "status_code" {
        value = step.function.hello_nodejs_step.result.statusCode
    }
}

pipeline "with_functions_no_env_var" {

    param "event" {
        type = any
    }

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

    step "function" "hello_nodejs_step" {
        runtime = "nodejs:18"
        handler = "index.handler"
        src = "./functions/hello-nodejs"
        event = param.event
    }

    output "val" {
        value = step.function.hello_nodejs_step.result.body.message
    }

    output "env" {
        value = step.function.hello_nodejs_step.result.body.env
    }

    output "status_code" {
        value = step.function.hello_nodejs_step.result.statusCode
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