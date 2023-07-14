pipeline "expr_depend_and_function" {
    step "text" "text_1" {
        text = "foo"
    }

    step "text" "text_2" {
        text = "lower case here ${title("bar ${step.text.text_1.text} baz")} lower case here again"
    }

    step "text" "text_3" {
        text = "output2 ${title(step.text.text_2.text)} title(output1) ${title(step.text.text_1.text)}"
    }
}