mod "simple_mod" {
    title = "Simple Mod"
}

pipeline "echo_one" {
    param "value" {
        type = string
        default = "Hello World"
    }

    step "transform" "echo_one" {
        value = "secret key: ${credential.aws.example.secret_key}"
    }

    output "val" {
        value = step.transform.echo_one.value
    }
}
