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