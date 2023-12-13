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

    param "test_run" {
        type = string
        default = "aye"
    }

    step "function" "hello_nodejs_step" {
        if = param.test_run == "aye"

        runtime = "nodejs:18"
        handler = "index.handler"
        src     = "./functions/hello-nodejs"
        event   = param.event
        timeout = 60000

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
        src     = "./functions/hello-nodejs"
        event   = param.event
        timeout = "60s"
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