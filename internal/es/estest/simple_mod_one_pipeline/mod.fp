mod "simple_mod" {
    title = "Simple Mod"
}

pipeline "echo_one" {
    param "value" {
        type = string
        default = var.var_one
    }

    param "value2" {
        type = list(string)
        default = ["Hello World", "Hello World 2"]
    }

    param "value3" {
        type = list(map(string))
    }

    param "value4" {
        type = map(list(number))
    }

    param "value5" {
        type = list(list(list(map(string))))
    }

    step "transform" "echo_one" {
        value = "secret key: ${credential.aws.example.secret_key}"
    }

    output "val" {
        value = step.transform.echo_one.value
    }
}
