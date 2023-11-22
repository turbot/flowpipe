pipeline "pipes_echo" {
    step "transform" "foo" {
        value = "foo"
    }

    output "foo" {
        value = step.transform.foo.value
    }
}


pipeline "pipes_list_echo" {

    param "string_list" {
        type = list(string)
    }

    step "transform" "my_string" {
        value = join(",", param.string_list)
    }

    output "foo" {
        value = step.transform.my_string.value
    }
}
