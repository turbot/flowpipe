pipeline "simple_text" {
    description = "text pipeline - debug should be removed"
    step "text" "text_1" {
        text = "foo"
    }
}