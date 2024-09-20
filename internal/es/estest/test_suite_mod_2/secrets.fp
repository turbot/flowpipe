pipeline "with_secrets" {
    param "conn" {
        type = connection.aws
        default = connection.aws.example
    }

    step "transform" "one" {
        value = param.conn
    }

    step "transform" "two" {
        value = step.transform.one.value.access_key
    }

    output "val" {
        value = step.transform.two.value
    }
}