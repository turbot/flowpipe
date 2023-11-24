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

    step "transform" "text_1" {
        value = "foo"
    }

    step "transform" "text_2" {
        value = "bar"
        if    = step.transform.text_1.value == "foo"
    }

    step "transform" "text_3" {
        value = "baz"
        if    = step.transform.text_1.value == "bar"
    }
}
