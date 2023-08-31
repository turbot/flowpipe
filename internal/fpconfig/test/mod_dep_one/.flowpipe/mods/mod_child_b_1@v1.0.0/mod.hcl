mod "mod_child_b" {
  title = "Child Mod B"
}

pipeline "this_pipeline_is_in_the_child" {
    step "echo" "foo" {
        text = "foo"
    }

    output "foo_a" {
        value = step.echo.foo.text
    }
}
