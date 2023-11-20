pipeline "pipes_echo" {
    step "echo" "foo" {
        text = "foo"
    }

    output "foo" {
        value = step.echo.foo.text
    }
}


pipeline "pipes_list_echo" {

    param "string_list" {
        type = list(string)
    }

    step "echo" "my_string" {
        text = join(",", param.string_list)
    }

    output "foo" {
        value = step.echo.my_string.text
    }
}
