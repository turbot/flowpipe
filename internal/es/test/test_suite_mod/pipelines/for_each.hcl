pipeline "for_each_empty_test" {

    param "input" {
        type = list(string)
        default = []
    }

    step "echo" "echo" {
        for_each = param.input
        text = each.value
    }


    output "echo" {
        value = step.echo.echo
    }
}

pipeline "for_each_non_collection" {

    param "input" {
        type = string
        default = "foo"
    }

    step "echo" "echo" {
        for_each = param.input
        text = each.value
    }


    output "echo" {
        value = step.echo.echo
    }
}