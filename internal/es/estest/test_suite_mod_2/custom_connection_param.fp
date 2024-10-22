

variable "connection" {
    type = connection
}

pipeline "steampipe_conn" {

    step "transform" "one" {
        value = connection.steampipe.default.host
    }

    output "val" {
        value = step.transform.one.value
    }
}

pipeline "steampipe_conn_with_param" {

    param "connection" {
        type = connection.steampipe
        default = connection.steampipe.default
    }

    step "transform" "one" {
        value = param.connection.host
    }

    output "val" {
        value = step.transform.one
    }
}

pipeline "connection_param_parent" {

    param "connection" {
        type = connection
        default = connection.steampipe.default
    }

    step "pipeline" "call_child" {
        pipeline = pipeline.connection_param_child
        args = {
            child_connection = param.connection
        }
    }

   output "val" {
        value = step.pipeline.call_child.output.val
    }
}
pipeline "connection_param_child" {

    param "child_connection" {
        type = connection
    }

    step "transform" "t" {
        value = param.child_connection.host
    }

    output "val" {
        value = step.transform.t
    }
}

pipeline "connection_var_param" {

    param "connection" {
        type = connection
        default = var.connection
    }

    step "transform" "connection" {
        value = param.connection.host
    }

    output "val" {
        value = step.transform.connection
    }
}

pipeline "connection_var" {

    step "transform" "connection" {
        value = var.connection.host
    }

    output "val" {
        value = step.transform.connection
    }
}