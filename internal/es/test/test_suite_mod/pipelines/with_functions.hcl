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
        }
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
        default = "AKIAQGDRKHTKBKCJASUB"
    }
    param "aws_secret_access_key" {
        type = string
        default = "N+rkACqwzo8gNQi4oxwJ14wYYIVmE2/jMoZ/XTzn"
    }

    step "container" "container_run" {
        image = "steampipe-aws-compliance"
        cmd = ["steampipe", "check", "aws_compliance.benchmark.audit_manager_control_tower_disallow_instances_5_1_1", "--output", "json"]
        env = {
            AWS_REGION = "us-east-1"
            AWS_ACCESS_KEY_ID = "AKIAQGDRKHTKBKCJASUB"
            AWS_SECRET_ACCESS_KEY = "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"
        }
    }
    // output "stdout" {
    //     value = jsondecode(step.container.container_run.stdout).groups[0].controls[0].results[3].resource
    // }
    // output "stderr" {
    //     value = step.container.container_run.stderr
    // }

    // step "echo" "element_bucket" {
    //     text = element(split(":", jsondecode(step.container.container_run.stdout).groups[0].controls[0].results[3].resource), 5)
    // }


    // output "echo_out1" {
    //     value = step.echo.element_bucket
    // }

    // step "echo" "container_run_aws1" {
    //     for_each = { for k,v in jsondecode(step.container.container_run.stdout).groups[0].controls[0].results  : k => v if v.status == "alarm"}
    //     // for_each = jsondecode(step.container.container_run.stdout).groups[0].controls[0].results
    //     // if = each.value.status == "alarm"
    //     // image = "amazon/aws-cli"
    //     text =  element(split(":", each.value.resource), 5)
    // }
    // output "test_echo" {
    //     value = step.echo.container_run_aws1
    // }
    step "container" "container_run_aws" {
        for_each = { for k,v in jsondecode(step.container.container_run.stdout).groups[0].controls[0].results  : k => v if v.status == "alarm"}
        image = "amazon/aws-cli"
        cmd = ["s3api", "put-bucket-versioning", "--bucket", element(split(":", each.value.resource), 5), "--versioning-configuration", "Status=Enabled"]
        env = {
            AWS_REGION = "us-east-1"
            AWS_ACCESS_KEY_ID = "AKIAQGDRKHTKBKCJASUB"
            AWS_SECRET_ACCESS_KEY = "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"
        }
    }

    output "stdout_aws" {
        value = step.container.container_run_aws[*].stdout
    }

   
    //  output "stderr_aws" {
    //     value = step.container.container_run_aws.stderr
    // }
}