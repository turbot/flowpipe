mod "simple_mod" {
    title = "Simple Mod"
}

pipeline "echo_one" {
    param "my_name" {
        type = string
        default = var.var_one
    }

    param "my_list" {
        type = list(string)
        default = ["Hello World", "Hello World 2"]
    }

    param "my_map" {
        type = map(string)
        default = {
            key1 = "value1"
            key2 = "value2"
        }
    }

    param "my_list_map" {
        type = list(map(string))
        default = [
            {
                key1 = "value1"
                key2 = "value2"
            },
            {
                key3 = "value3"
                key4 = "value4"
            }
        ]
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
