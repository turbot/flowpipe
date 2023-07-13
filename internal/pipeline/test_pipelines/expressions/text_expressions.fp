pipeline "text_expr" {
    step "text" "text_1" {
        text = "foo"
    }

    step "text" "text_2" {
        text = "bar ${step.text.text_1.text} baz"
    }

    step "text" "text_3" {
        text = "bar ${step.text.text_2.text} baz ${step.text.text_1.text}"
    }
}