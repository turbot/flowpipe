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
        default = ["jerry", "Janis", "Jimi"]
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