pipeline "steampipe_conn" {

    step "transform" "one" {
        value = connection.steampipe.default.connection_string
    }

    output "val" {
        value = step.transform.one.value
    }
}

pipeline "steampipe_conn_with_param" {

    param "conn_name" {
        type = connection.steampipe
        default = connection.steampipe.default
    }

    step "transform" "one" {
        value = param.conn_name.connection_string
    }

    output "val" {
        value = step.transform.one.value
    }
}