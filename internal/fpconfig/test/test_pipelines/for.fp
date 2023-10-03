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

pipeline "for_depend_object" {

    param "users" {
        type = list(string)
        default = ["brian", "freddie", "john", "roger"]
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


pipeline "for_loop_depend" {

    param "users" {
        type = list(string)
        default = ["jerry", "Janis", "Jimi"]
    }

    step "echo" "text_1" {
        for_each = param.users
        text = "user is ${each.value}"
    }

    step "echo" "text_2" {
        text = "output is ${step.echo.text_1[0].text}"
    }

    step "echo" "text_3" {
        for_each = step.echo.text_1
        text = "output one value is ${each.value.text}"
    }
}


pipeline "for_map" {
    param "legends" {
        type = map

        default = {
            "janis" = {
                last_name= "joplin"
                age = 27
            }
            "jimi" = {
                last_name= "hendrix"
                age = 27
            }
            "jerry" = {
                last_name= "garcia"
                age = 53
            }
        }
    }

    step "echo" "text_1" {
        for_each = param.legends
        text = "${each.value.key} ${each.value.last_name} was ${each.value.age}"
    }
}