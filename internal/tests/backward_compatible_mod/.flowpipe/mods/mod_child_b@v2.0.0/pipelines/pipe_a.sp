pipeline "second_pipe_in_the_child" {
    step "echo" "foo" {
        text = "foo"
    }

    output "foo_b" {
        value = step.echo.foo.text
    }
}
