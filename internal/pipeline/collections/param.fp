pipeline "param_text" {

    param "text" {
        default = "foo"
    }

    step "echo" "text_1" {
        text = param.text
    }
}