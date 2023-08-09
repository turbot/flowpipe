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

pipeline "for_static" {
    step "echo" "text_1" {
        for_each = ["one", "two", "three"]
        text = "user if ${each.value}"
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

pipeline "for_loop_with_if" {

    param "time" {
        type = list(string)
        default = ["1s", "2s", "3s"]
    }

    step "sleep" "sleep_1" {
        for_each = param.time
        duration = each.value
    }

    step "echo" "echo_1" {
        for_each = step.sleep.sleep_1
        text = "sleep 1 output: ${each.value.duration}"
        if = each.value.duration == "1s"
    }
}

pipeline "for_rows_in_http" {
    step "http" "my_step_1" {
        url = "https://jsonplaceholder.typicode.com/posts"
        method = "Post"
        request_body = jsonencode({
            userId = 12345
            users = [
                {
                    name = "billy joe armstrong"
                },
                {
                    name = "mike dirnt"
                },
                {
                    name = "tre cool"
                }
            ]
        })
        request_headers = {
            Accept = "*/*"
            Content-Type = "application/json"
            User-Agent = "flowpipe"
        }
        request_timeout_ms = 3000
    }

    step "echo" "extract_users" {
       text = jsonencode(step.http.my_step_1.response_body)
       // text = "${ join("", [for row in jsondecode(step.http.my_step_1.response_body): "\n- ${row.name}"]) }"
    }
}

pipeline "list_for" {
    param "user_data" {
        type = map(list(string))
        default = {
            Users = ["jim", "jeff", "jerry"]
        }
    }

    step "echo" "example" {
        for_each = { for user in param.user_data.Users : user => user }
        # Other resource attributes here, if needed
        text = each.value
    }
}