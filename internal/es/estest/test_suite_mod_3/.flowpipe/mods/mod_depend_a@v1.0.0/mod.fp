mod "mod_depend_a" {
    title = "Child mod A"

    require {
        mod "mod_depend_b" {
            version = "1.0.0"
        }
    }
}

pipeline "bar_a" {
    step "transform" "bar" {
        value = "bar"
    }

    output "val" {
        value = step.transform.bar.value
    }
}


pipeline "optional_message_a" {
    param "notifier" {
        type = string
        default = "foo"
    }

    param "send" {
        type = bool
        default = true
    }

    param "text" {
        type = string
        default = "Hello World"
    }

    step "transform" "output" {
        value = "${param.text} - ${param.notifier} - ${param.send}"
    }

    output "val" {
        value = step.transform.output.value
    }
}

pipeline "enforce_a" {
    param "notifier" {
        type = string
        default = "enforce"
    }

    param "send" {
        type = bool
        default = true
    }

    param "text" {
        type = string
        default = "Enforce Pipeline"
    }

    step "transform" "output" {
        value = "${param.text} - ${param.notifier} - ${param.send}"
    }

    output "val" {
        value = step.transform.output.value
    }
}
