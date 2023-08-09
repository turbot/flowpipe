pipeline "with_output" {
    step "echo" "echo_1" {
        text = "foo bar"
    }

    output "one" {
        value = step.echo.echo_1.text
    }

    output "two" {
        value = title(step.echo.echo_1.text)
    }
}