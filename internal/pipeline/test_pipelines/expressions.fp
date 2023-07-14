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

pipeline "expr_func" {
    step "text" "text_title" {
        text = title("Hello World")
    }
}

pipeline "expr_within_text" {
    step "text" "text_title" {
        text = "Hello ${title("world")}"
    }
}


pipeline "expr_depend_and_function" {
    step "text" "text_1" {
        text = "foo"
    }

    step "text" "text_2" {
        text = title("bar ${step.text.text_1.text} baz")
    }

    step "text" "text_3" {
        text = "output2 ${title(step.text.text_2.text)} func(output1) ${func(step.text.text_1.text)}"
    }
}