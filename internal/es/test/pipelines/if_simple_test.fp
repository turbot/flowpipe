pipeline "simple_if" {
    // param "condition_true" {
    //     type = bool
    //     default = true
    // }

    // step "echo" "text_true" {
    //     text = "foo"
    //     if = param.condition_true
    // }

    // param "condition_false" {
    //     type = bool
    //     default = false
    // }

    // step "echo" "text_false" {
    //     text = "foo"
    //     if = param.condition_false
    // }

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
