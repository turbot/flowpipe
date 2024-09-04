pipeline  "dynamic_conn_parent" {

    step "pipeline" "call_dynamic_conn" {
        pipeline = pipeline.dynamic_conn

        // simulate acquiring multiple connections from a query step
        for_each = ["sso", "dundermifflin"]

        args = {
            conn = each.value
        }
    }

    output "val_0" {
        value = step.pipeline.call_dynamic_conn[0].output.val
    }

    output "val_1" {
        value = step.pipeline.call_dynamic_conn[1].output.val
    }
}

pipeline "dynamic_conn" {

    param "conn" {
        type = string
        default = "example"
    }

    step "transform" "test" {
        output "val" {
            value = connection.aws[param.conn]
        }
    }

    output "val" {
        value = step.transform.test.output.val.access_key
    }
}
