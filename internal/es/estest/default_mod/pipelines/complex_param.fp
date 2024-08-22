pipeline "complex_param" {
    param "my_name" {
        type = string
        default = "hello"
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

    param "actions" {
        description = "A map of actions, if approvers are set these will be offered as options to select, else the one matching the default_action will be used."
        type = map(object({
        label         = string
        value         = string
        style         = string
        pipeline_args = any
        success_msg   = string
        error_msg     = string
        }))
        default = {
        "skip" = {
            label         = "Skip"
            value         = "skip"
            style         = "info"
            pipeline_args = {
            notifier = "default"
            send     = false
            text     = "Skipped item."
            }
            success_msg   = ""
            error_msg     = ""
        }
        }
    }

    step "transform" "echo_one" {
        value = "secret key: ${credential.aws.example.secret_key}"
    }

    output "val" {
        value = step.transform.echo_one.value
    }
}
