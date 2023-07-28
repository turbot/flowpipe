pipeline "if" {
    param "condition" {
        type = bool
        default = true
    }

    step "echo" "text_1" {
        text = "foo"
        if = param.condition
    }
}

pipeline "if_negative" {
    param "condition" {
        type = bool
        default = false
    }

    step "echo" "text_1" {
        text = "foo"
        if = param.condition
    }
}

pipeline "if_depends" {
    step "echo" "text_1" {
        text = "foo"
    }

    step "echo" "text_2" {
        text = "bar"
        if = step.echo.text_1.text == "foo"
    }

    step "echo" "text_3" {
        text = "baz"
        if = step.echo.text_1.text == "bar"
    }
}