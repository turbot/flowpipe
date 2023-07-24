pipeline "text_expr" {
    step "echo" "text_1" {
        text = "foo"
    }

    step "echo" "text_2" {
        text = "bar ${step.echo.text_1.text} baz"
    }

    step "echo" "text_3" {
        text = "bar ${step.echo.text_2.text} baz ${step.echo.text_1.text}"
    }
}

pipeline "expr_func" {
    step "echo" "text_title" {
        text = title("Hello World")
    }
}

pipeline "expr_within_text" {
    step "echo" "text_title" {
        text = "Hello ${title("world")}"
    }
}


pipeline "expr_depend_and_function" {
    step "echo" "text_1" {
        text = "foo"
    }

    step "echo" "text_2" {
        text = title("bar ${step.echo.text_1.text} baz")
    }

    step "echo" "text_3" {
        text = "output2 ${title(step.echo.text_2.text)} func(output1) ${func(step.echo.text_1.text)}"
    }
}