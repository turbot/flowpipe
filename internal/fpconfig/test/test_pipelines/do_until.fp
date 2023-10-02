pipeline "do_until" {
    step "echo" "repeat" {
        text  = "iteration no"
        numeric = 5
    }

    output "echo" {
        value = step.echo.repeat.numeric
    }
}