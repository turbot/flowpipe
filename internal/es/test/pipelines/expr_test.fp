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

    step "echo" "explicit_depends" {
        depends_on = [
            step.echo.text_2,
            step.echo.text_1
        ]
        text = "explicit depends here"
    }

    # "time"/"for"/"sleep" steps
     param "time" {
        type = list(string)
        default = ["1s", "2s"]
    }

    step "sleep" "sleep_1" {
        for_each = param.time
        duration = each.value
    }

    step "echo" "echo_sleep_for" {
        for_each = step.sleep.sleep_1
        text = each.value.duration
    }

    step "echo" "echo_sleep_1" {
        text = "sleep 2 output: ${step.echo.echo_sleep_for[1].text}"
    }

    step "echo" "echo_sleep_2" {
        text = "sleep 1 output: ${step.sleep.sleep_1[0].duration}"
    }
}