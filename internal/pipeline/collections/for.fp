pipeline "for_loop" {

    param "users" {
        type = list(string)
        default = ["jerry", "Janis", "Jimi"]
    }

    step "echo" "text_1" {
        for_each = param.users
        text = "user if ${each.value}"
    }

    step "echo" "no_for_each" {
        text = "baz"
    }
}

pipeline "for_loop_nested_with_sleep" {

    param "time" {
        type = list(string)
        default = ["1s", "2s"]
    }

    step "sleep" "sleep_1" {
        for_each = param.time
        duration = each.value
    }

    step "sleep" "sleep_2" {
        for_each = step.sleep.sleep_1
        duration = each.value.duration
    }
}

pipeline "for_loop_nested_with_sleep_and_index" {

    param "time" {
        type = list(string)
        default = ["1s", "2s"]
    }

    step "sleep" "sleep_1" {
        for_each = param.time
        duration = each.value
    }

    step "sleep" "sleep_2" {
        for_each = step.sleep.sleep_1
        duration = each.value.duration
    }

    step "echo" "echo_1" {
        text = "sleep 2 output: ${step.sleep.sleep_2[1].duration}"
    }

    step "echo" "echo_2" {
        text = "sleep 1 output: ${step.sleep.sleep_1[0].duration}"
    }
}

pipeline "for_loop_nested" {

    param "users" {
        type = list(string)
        default = ["brian", "freddie"]
    }

    step "echo" "text_1" {
        for_each = param.users
        text = "user is ${each.value}"
    }

    step "echo" "text_2" {
        for_each = step.echo.text_1
        text = "output is ${each.value.text}"
    }
}

pipeline "for_depend_object" {

    param "users" {
        type = list(string)
        default = ["freddie", "brian"]
    }

    step "echo" "text_1" {
        for_each = param.users
        text = "user if ${each.value}"
    }

    step "echo" "text_3" {
        for_each = step.echo.text_1
        text = "output one value is ${each.value.text}"
    }
}


pipeline "for_loop_depends" {

    param "users" {
        type = list(string)
        default = ["jerry", "Janis", "Jimi"]
    }

    step "echo" "text_1" {
        for_each = param.users
        text = "user is ${each.value}"
    }

    step "echo" "text_2" {
        // text = output is user is jerry
        text = "output is ${step.echo.text_1[0].text}"
    }

    step "echo" "text_3" {
        for_each = step.echo.text_1
        text = "output one value is ${each.value.text}"
    }
}