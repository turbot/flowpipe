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

pipeline "container_with_for_each" {
    param "input" {
        type = list(any)
        default = [
            {
                name = "foo"
                value = "bar"
            },
            {
                name = "baz"
                value = "qux"
            },
            {
                name = "quux"
                value = "quuz"
            }
        ]
    }

    step "container" "container_run_aws" {
        for_each = param.input
        image = "amazon/aws-cli"
        cmd = ["s3api", "put-bucket-versioning", "--bucket", element(split(":", each.value.name), 5), "--versioning-configuration", "Status=Enabled"]
        env = {
            AWS_REGION = "us-east-1"
            AWS_ACCESS_KEY_ID = "AKIAQGDRKHTKBKCJASUB"
            AWS_SECRET_ACCESS_KEY = "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"
        }
    }


    step "echo" "foo" {
        for_each = param.input
        text = "echo ${each.value.name}"
    }

    step "echo" "baz" {
        if = false
        text = "not here"
    }

    output "baz" {
        value = step.echo.baz.text
    }

    output "val" {
        value = step.echo.foo
    }
}