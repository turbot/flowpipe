pipeline "pipes_echo" {
    step "echo" "foo" {
        text = "foo"
    }

    output "foo" {
        value = step.echo.foo.text
    }
}
