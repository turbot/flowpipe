pipeline "parent_pipeline_sp_nested" {
    step "echo" "foo" {
        text = "foo"
    }

    output "foo_b" {
        value = step.echo.foo.text
    }
}
