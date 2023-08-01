pipeline "expr_depend_and_function" {
    step "echo" "text_1" {
        text = "foo bar"
    }

    step "echo" "text_2" {
        text = "lower case ${title("bar ${step.echo.text_1.text} baz")} and here"
    }

    step "echo" "text_3" {
        text = "output 2 ${title(step.echo.text_2.text)} title(output1) ${title(step.echo.text_1.text)}"
    }
}