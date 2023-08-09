pipeline "simple_text" {
    description = "text pipeline - debug should be removed"
    step "echo" "text_1" {
        text = "foo"
    }
}

