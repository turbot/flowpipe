pipeline "do_until" {
    step "echo" "repeat" {
        text  = "iteration no"
        numeric = 5
    }

    step "echo" "repeat_two" {
        text  = "iteration no"

        do_until = last.output.token == ""
        do_until = last.index > 4

        numeric = last.index
    }

    output "echo" {
        value = step.echo.repeat.numeric
    }

    output "echo_two" {
        value = step.echo.repeat_two.numeric
    }
}