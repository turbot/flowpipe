pipeline "parent_with_creds" {

    param "cred" {
        type = string
        default = "example_two"
    }

    step "pipeline" "call_child" {
        pipeline = pipeline.nested_with_creds
        args = {
            cred = param.cred
        }
    }

    output "env" {
        value = step.pipeline.call_child.output.env
    }

    output "env_2" {
        value = step.pipeline.call_child.output.env_2

    }
}

pipeline "nested_with_creds" {
    param "cred" {
        type = string
    }

    step "transform" "echo" {
        value = credential.aws[param.cred].env
    }

    step "transform" "with_merge" {
        value = merge(credential.aws[param.cred].env, { AWS_REGION = param.cred })
    }

    output "env" {
        value = step.transform.echo.value
    }

    output "env_2" {
        value = step.transform.with_merge.value
    }
}

pipeline "parent_call_nested_mod_with_cred" {

    step "pipeline" "call_child" {
        pipeline = mod_depend_a.pipeline.with_github_creds
    }

    output "val" {
        value = step.pipeline.call_child.output.val
    }

    output "val_merge" {
        value = step.pipeline.call_child.output.val_merge
    }
}