pipeline "simple_text" {
    description = "text pipeline - debug should be removed"
    step "echo" "text_1" {
        text = "foo"
    }
}

pipeline "simple_list" {
    description = "text pipeline - debug should be removed"
    step "echo" "text_1" {
        list_text = ["foo", "bar", "baz"]
    }
}