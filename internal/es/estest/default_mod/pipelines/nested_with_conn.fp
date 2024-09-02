pipeline "parent_with_conn" {

    param "conn" {
        type = string
        default = "example_two"
    }

    step "pipeline" "call_child" {
        pipeline = pipeline.nested_with_conn
        args = {
            conn = param.conn
        }
    }

    output "env" {
        value = step.pipeline.call_child.output.env
    }

    output "env_2" {
        value = step.pipeline.call_child.output.env_2

    }
}

pipeline "nested_with_conn" {
    param "conn" {
        type = string
    }

    step "transform" "echo" {
        value = connection.aws[param.conn].env
    }

    step "transform" "with_merge" {
        value = merge(connection.aws[param.conn].env, { AWS_REGION = param.conn })
    }

    output "env" {
        value = step.transform.echo.value
    }

    output "env_2" {
        value = step.transform.with_merge.value
    }
}

pipeline "parent_call_nested_mod_with_conn" {

    step "pipeline" "call_child" {
        pipeline = mod_depend_a.pipeline.with_github_conns
        args = {
            conn = "default"
        }
    }

    output "val" {
        value = step.pipeline.call_child.output.val
    }

    output "val_merge" {
        value = step.pipeline.call_child.output.val_merge
    }
}

pipeline "parent_call_nested_mod_with_conn_with_invalid_conn" {

    // step "input" "test_input" {
    //     notifier = notifier.default
    //     prompt = "Do you want to approve?"
    //     type = "button"

    //     option "Approve" {}
    //     option "Deny" {}
    // }

    step "pipeline" "call_child" {
        pipeline = mod_depend_a.pipeline.with_github_conns
        args = {
            conn = "my_github_creds"
        }
    }

    output "val" {
        value = step.pipeline.call_child.output.val
    }

    output "val_merge" {
        value = step.pipeline.call_child.output.val_merge
    }
}

