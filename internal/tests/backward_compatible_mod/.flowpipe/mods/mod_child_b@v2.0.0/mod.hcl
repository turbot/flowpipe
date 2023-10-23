mod "mod_child_b" {
  title = "Child Mod b"
}

pipeline "this_pipeline_is_in_the_child" {
    step "echo" "foo" {
        text = "foo"
    }

    output "foo_b" {
        value = step.echo.foo.text
    }
}
